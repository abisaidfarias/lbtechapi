// migrate_shipment_verdict backfills the verdict of historical shipment
// controls: every control already closed (current_phase 3) without a verdict
// becomes "approved", which is what the old "Completed" status meant.
// Partial approvals need nothing extra: partial is derived from the counts.
//
// Only documents without a verdict are matched, so re-running is a no-op and
// a verdict set by the new flow (approved or rejected) is never overwritten.
//
// Usage (from the repo root, which holds .env with MONGO_URI / MONGO_DB):
//
//	go run ./tools/migrate_shipment_verdict            # dry run: counts only, writes nothing
//	go run ./tools/migrate_shipment_verdict -apply     # backup + migrate
//
// Flags: -apply, -backup <file> (default: OS temp dir), -env <file> (default .env).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionName  = "shipment_controls"
	phaseCompleted  = 3
	verdictApproved = "approved"
)

// pendingFilter matches closed controls with no verdict (missing, null or "").
func pendingFilter() bson.M {
	return bson.M{
		"current_phase": phaseCompleted,
		"$or": bson.A{
			bson.M{"verdict": bson.M{"$exists": false}},
			bson.M{"verdict": nil},
			bson.M{"verdict": ""},
		},
	}
}

func migrationUpdate() bson.M {
	return bson.M{"$set": bson.M{"verdict": verdictApproved}}
}

func main() {
	apply := flag.Bool("apply", false, "write changes (default is a dry run)")
	envFile := flag.String("env", ".env", "env file with MONGO_URI and MONGO_DB")
	backupPath := flag.String("backup", "", "backup file (default: OS temp dir)")
	flag.Parse()

	_ = godotenv.Load(*envFile)
	uri, dbName := os.Getenv("MONGO_URI"), os.Getenv("MONGO_DB")
	if uri == "" || dbName == "" {
		fmt.Fprintln(os.Stderr, "MONGO_URI and MONGO_DB are required (env or -env file)")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	coll := client.Database(dbName).Collection(collectionName)

	mode := "DRY RUN (use -apply to write)"
	if *apply {
		mode = "APPLY"
	}
	fmt.Printf("migrate_shipment_verdict on database %q: %s\n", dbName, mode)

	total, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	closed, err := coll.CountDocuments(ctx, bson.M{"current_phase": phaseCompleted})
	if err != nil {
		panic(err)
	}
	pending, err := coll.CountDocuments(ctx, pendingFilter())
	if err != nil {
		panic(err)
	}
	partial, err := coll.CountDocuments(ctx, bson.M{
		"$and": bson.A{
			pendingFilter(),
			bson.M{"$expr": bson.M{"$lt": bson.A{"$registered_imei_count", "$imei_quantity"}}},
		},
	})
	if err != nil {
		panic(err)
	}
	zeroRegistered, err := coll.CountDocuments(ctx, bson.M{
		"$and": bson.A{pendingFilter(), bson.M{"registered_imei_count": bson.M{"$lte": 0}}},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("  total=%d closed(phase 3)=%d pending_without_verdict=%d (of which partial=%d, with 0 registered=%d)\n",
		total, closed, pending, partial, zeroRegistered)

	if !*apply || pending == 0 {
		if pending == 0 {
			fmt.Println("nothing to migrate (already migrated)")
		}
		return
	}

	if *backupPath == "" {
		*backupPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("migrate_shipment_verdict_backup_%s_%s.jsonl", dbName, time.Now().Format("20060102_150405")))
	}
	f, err := os.Create(*backupPath)
	if err != nil {
		panic(err)
	}
	cur, err := coll.Find(ctx, pendingFilter(), options.Find().SetProjection(bson.M{
		"verdict": 1, "current_phase": 1, "status": 1, "registered_imei_count": 1, "imei_quantity": 1,
	}))
	if err != nil {
		panic(err)
	}
	for cur.Next(ctx) {
		line, err := bson.MarshalExtJSON(cur.Current, true, false)
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
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("backup written to", *backupPath)

	res, err := coll.UpdateMany(ctx, pendingFilter(), migrationUpdate())
	if err != nil {
		panic(err)
	}
	left, err := coll.CountDocuments(ctx, pendingFilter())
	if err != nil {
		panic(err)
	}
	fmt.Printf("  %s matched=%d modified=%d remaining_without_verdict=%d\n",
		collectionName, res.MatchedCount, res.ModifiedCount, left)
	if left > 0 {
		os.Exit(1)
	}
	fmt.Println("done")
}
