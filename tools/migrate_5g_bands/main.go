// migrate_5g_bands splits the legacy 5G band lists into 5G NSA and 5G SA.
//
// Existing values move to NSA (user decision); SA starts empty:
//
//	configurations:   bands_5g -> bands_5g_nsa, bands_5g_sa = []
//	countries:        band_5g  -> band_5g_nsa,  band_5g_sa  = []
//	devices:          band_5g  -> band_5g_nsa,  band_5g_sa  = []
//	device_trackings: tracking_logs[].country.band_5g (country snapshot) -> same split
//
// Each document is rewritten by a single atomic pipeline update that copies the
// legacy list into *_nsa and removes the legacy field. Only documents that still
// have the legacy field are matched, so re-running is a no-op; if *_nsa already
// exists, legacy values are appended without duplicates.
//
// Usage (from the repo root, which holds .env with MONGO_URI / MONGO_DB):
//
//	go run ./tools/migrate_5g_bands            # dry run: counts only, writes nothing
//	go run ./tools/migrate_5g_bands -preview   # read-only: one sample per collection before/after
//	go run ./tools/migrate_5g_bands -apply     # backup + migrate
//
// Flags: -apply, -preview, -backup <file> (default: OS temp dir), -env <file> (default .env).
// Environment variables MONGO_URI / MONGO_DB are used when set and no .env value exists.
// Requires MongoDB 4.2+ (pipeline updates).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type topLevelSplit struct {
	collection, legacy, nsa, sa string
}

var topLevel = []topLevelSplit{
	{"configurations", "bands_5g", "bands_5g_nsa", "bands_5g_sa"},
	{"countries", "band_5g", "band_5g_nsa", "band_5g_sa"},
	{"devices", "band_5g", "band_5g_nsa", "band_5g_sa"},
}

const (
	trackingCollection = "device_trackings"
	trackingLegacyPath = "tracking_logs.country.band_5g"
)

// mergeNsa returns ifNull(nsa, []) followed by the legacy values not already in it.
func mergeNsa(nsaExpr, legacyExpr string) bson.M {
	existing := bson.M{"$ifNull": bson.A{nsaExpr, bson.A{}}}
	legacy := bson.M{"$cond": bson.A{bson.M{"$isArray": legacyExpr}, legacyExpr, bson.A{}}}
	return bson.M{"$concatArrays": bson.A{
		existing,
		bson.M{"$filter": bson.M{
			"input": legacy,
			"as":    "b",
			"cond":  bson.M{"$not": bson.A{bson.M{"$in": bson.A{"$$b", existing}}}},
		}},
	}}
}

func topLevelPipeline(s topLevelSplit) mongo.Pipeline {
	return mongo.Pipeline{
		{{Key: "$set", Value: bson.D{
			{Key: s.nsa, Value: mergeNsa("$"+s.nsa, "$"+s.legacy)},
			{Key: s.sa, Value: bson.M{"$ifNull": bson.A{"$" + s.sa, bson.A{}}}},
		}}},
		{{Key: "$unset", Value: s.legacy}},
	}
}

func trackingPipeline() mongo.Pipeline {
	splitCountry := bson.M{"$mergeObjects": bson.A{
		"$$tl.country",
		bson.M{
			"band_5g_nsa": mergeNsa("$$tl.country.band_5g_nsa", "$$tl.country.band_5g"),
			"band_5g_sa":  bson.M{"$ifNull": bson.A{"$$tl.country.band_5g_sa", bson.A{}}},
		},
	}}
	hasLegacy := bson.M{"$ne": bson.A{bson.M{"$type": "$$tl.country.band_5g"}, "missing"}}
	return mongo.Pipeline{
		{{Key: "$set", Value: bson.M{
			"tracking_logs": bson.M{"$map": bson.M{
				"input": "$tracking_logs",
				"as":    "tl",
				"in": bson.M{"$cond": bson.A{
					hasLegacy,
					bson.M{"$mergeObjects": bson.A{"$$tl", bson.M{"country": splitCountry}}},
					"$$tl",
				}},
			}},
		}}},
		// Exclusion paths traverse arrays, so this drops country.band_5g in every log.
		{{Key: "$unset", Value: trackingLegacyPath}},
	}
}

// previewFields keeps only the 5G band fields (top level or inside tracking logs).
func previewFields(doc bson.M) bson.M {
	out := bson.M{"_id": doc["_id"]}
	for k, v := range doc {
		if strings.Contains(k, "5g") {
			out[k] = v
		}
	}
	if logs, ok := doc["tracking_logs"].(bson.A); ok {
		var countries bson.A
		for _, l := range logs {
			if m, ok := l.(bson.M); ok {
				if c, ok := m["country"].(bson.M); ok {
					countries = append(countries, previewFields(c))
				}
			}
		}
		out["tracking_logs.country"] = countries
	}
	return out
}

func main() {
	apply := flag.Bool("apply", false, "write changes (default is a dry run)")
	envFile := flag.String("env", ".env", "env file with MONGO_URI and MONGO_DB")
	backupPath := flag.String("backup", "", "backup file (default: OS temp dir)")
	preview := flag.Bool("preview", false, "read-only: show one migrated document per collection")
	flag.Parse()

	_ = godotenv.Load(*envFile)
	uri, dbName := os.Getenv("MONGO_URI"), os.Getenv("MONGO_DB")
	if uri == "" || dbName == "" {
		fmt.Fprintln(os.Stderr, "MONGO_URI and MONGO_DB are required (env or -env file)")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	db := client.Database(dbName)

	mode := "DRY RUN (use -apply to write)"
	if *apply {
		mode = "APPLY"
	}
	fmt.Printf("migrate_5g_bands on database %q: %s\n", dbName, mode)

	type job struct {
		collection, legacyPath string
		pipeline               mongo.Pipeline
	}
	var jobs []job
	for _, s := range topLevel {
		jobs = append(jobs, job{s.collection, s.legacy, topLevelPipeline(s)})
	}
	jobs = append(jobs, job{trackingCollection, trackingLegacyPath, trackingPipeline()})

	pending := 0
	for _, j := range jobs {
		n, err := db.Collection(j.collection).CountDocuments(ctx, bson.M{j.legacyPath: bson.M{"$exists": true}})
		if err != nil {
			panic(err)
		}
		pending += int(n)
		fmt.Printf("  %-17s %-30s pending=%d\n", j.collection, j.legacyPath, n)
	}
	if *preview {
		// Read-only: run the same stages as an aggregation on one pending
		// document per collection and print the legacy and split fields.
		for _, j := range jobs {
			stages := append(mongo.Pipeline{
				{{Key: "$match", Value: bson.M{j.legacyPath: bson.M{"$exists": true}}}},
				{{Key: "$limit", Value: 1}},
			}, j.pipeline...)
			before, err := db.Collection(j.collection).Aggregate(ctx, stages[:2])
			if err != nil {
				panic(err)
			}
			after, err := db.Collection(j.collection).Aggregate(ctx, stages)
			if err != nil {
				panic(err)
			}
			var b, a []bson.M
			_ = before.All(ctx, &b)
			_ = after.All(ctx, &a)
			if len(a) == 0 {
				continue
			}
			fmt.Printf("\n[%s] before: %v\n[%s] after:  %v\n",
				j.collection, previewFields(b[0]), j.collection, previewFields(a[0]))
		}
		return
	}
	if !*apply || pending == 0 {
		if pending == 0 {
			fmt.Println("nothing to migrate (already migrated)")
		}
		return
	}

	// Backup the legacy values of every pending document before touching anything.
	if *backupPath == "" {
		*backupPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("migrate_5g_bands_backup_%s_%s.jsonl", dbName, time.Now().Format("20060102_150405")))
	}
	f, err := os.Create(*backupPath)
	if err != nil {
		panic(err)
	}
	for _, j := range jobs {
		filter := bson.M{j.legacyPath: bson.M{"$exists": true}}
		cur, err := db.Collection(j.collection).Find(ctx, filter,
			options.Find().SetProjection(bson.M{j.legacyPath: 1}))
		if err != nil {
			panic(err)
		}
		for cur.Next(ctx) {
			line, err := bson.MarshalExtJSON(bson.M{"collection": j.collection, "doc": cur.Current}, true, false)
			if err != nil {
				panic(err)
			}
			if _, err := f.Write(append(line, '\n')); err != nil {
				panic(err)
			}
		}
		if err := cur.Err(); err != nil {
			panic(err)
		}
		cur.Close(ctx)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("backup written to", *backupPath)

	var failed []string
	for _, j := range jobs {
		coll := db.Collection(j.collection)
		filter := bson.M{j.legacyPath: bson.M{"$exists": true}}
		res, err := coll.UpdateMany(ctx, filter, j.pipeline)
		if err != nil {
			panic(fmt.Errorf("%s: %w", j.collection, err))
		}
		left, err := coll.CountDocuments(ctx, filter)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  %-17s matched=%d modified=%d remaining_legacy=%d\n",
			j.collection, res.MatchedCount, res.ModifiedCount, left)
		if left > 0 {
			failed = append(failed, j.collection)
		}
	}
	if len(failed) > 0 {
		fmt.Fprintln(os.Stderr, "legacy fields remain in:", strings.Join(failed, ", "))
		os.Exit(1)
	}
	fmt.Println("done")
}
