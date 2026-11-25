// migrate_device_catalogs seeds device_catalog_values from the legacy
// configuration arrays (device_types, roms, rams, memory_capabilities,
// usb_communications), keeping their order, deduplicated (trimmed,
// case-insensitive) and active.
//
// Idempotent: a value already in the catalog (same category and key) is
// skipped, and new ones are appended after the current last order.
//
// It also REPORTS, without deleting anything:
//   - catalog values used by no device, flagging the ones that look like test
//     data;
//   - values devices use that are not in their catalog (e.g. values moved to
//     the wrong list by the old Settings bug).
//
// Usage (from the repo root, which holds .env with MONGO_URI / MONGO_DB):
//
//	go run ./tools/migrate_device_catalogs            # dry run + report
//	go run ./tools/migrate_device_catalogs -apply     # backup + seed + report
//
// Flags: -apply, -backup <file> (default: OS temp dir), -env <file> (default .env).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type category struct {
	key, label, deviceField, configField string
}

// Same list as enums.DeviceCatalogCategories (kept local so the tool has no
// dependency on the API packages and their database init).
var categories = []category{
	{"device_type", "Device type", "type", "device_types"},
	{"rom", "ROM", "memory_rom", "roms"},
	{"ram", "RAM", "memory_ram", "rams"},
	{"memory_capability", "Memory capabilities", "memory_cpu", "memory_capabilities"},
	{"usb", "USB", "communication_usb", "usb_communications"},
}

func valueKey(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

// seedValues returns the distinct non-empty values of a legacy array, in
// order (first occurrence wins).
func seedValues(raw []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range raw {
		v = strings.TrimSpace(v)
		k := valueKey(v)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, v)
	}
	return out
}

// isRepeatedChar reports values made of one repeated character ("rrr", "2222").
func isRepeatedChar(k string) bool {
	r := []rune(k)
	if len(r) < 2 {
		return false
	}
	for _, c := range r[1:] {
		if c != r[0] {
			return false
		}
	}
	return true
}

var (
	testWords = map[string]bool{
		"test": true, "prueba": true, "asd": true, "asdf": true, "qwe": true, "xxx": true,
		"cap": true, "memory": true, "rom": true, "ram": true, "usb": true, "type": true,
		"device": true, "value": true, "new": true, "nuevo": true, "123": true, "1234": true,
	}
)

// testLikeReason explains why an unused value looks like test data ("" when
// it does not).
func testLikeReason(v string) string {
	k := valueKey(v)
	switch {
	case testWords[k]:
		return "generic/test word"
	case isRepeatedChar(k):
		return "repeated character"
	case strings.Contains(k, "test") || strings.Contains(k, "prueba"):
		return "contains test/prueba"
	case len(k) <= 2 && !strings.ContainsAny(k, "0123456789"):
		return "too short"
	}
	return ""
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
	db := client.Database(dbName)
	catalog := db.Collection("device_catalog_values")

	mode := "DRY RUN (use -apply to write)"
	if *apply {
		mode = "APPLY"
	}
	fmt.Printf("migrate_device_catalogs on database %q: %s\n", dbName, mode)

	var config bson.M
	if err := db.Collection("configurations").FindOne(ctx, bson.M{}).Decode(&config); err != nil && err != mongo.ErrNoDocuments {
		panic(err)
	}

	type planned struct {
		cat   category
		value string
		order int
	}
	var plan []planned
	existingByCat := map[string][]bson.M{}
	for _, c := range categories {
		var existing []bson.M
		cur, err := catalog.Find(ctx, bson.M{"category": c.key}, options.Find().SetSort(bson.M{"order": 1}))
		if err != nil {
			panic(err)
		}
		_ = cur.All(ctx, &existing)
		existingByCat[c.key] = existing
		known := map[string]bool{}
		next := 0
		for _, e := range existing {
			known[valueKey(fmt.Sprint(e["value"]))] = true
			if o, ok := e["order"].(int32); ok && int(o) >= next {
				next = int(o) + 1
			}
		}
		var raw []string
		if arr, ok := config[c.configField].(bson.A); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					raw = append(raw, s)
				}
			}
		}
		toAdd := 0
		for _, v := range seedValues(raw) {
			if known[valueKey(v)] {
				continue
			}
			plan = append(plan, planned{c, v, next})
			next++
			toAdd++
		}
		fmt.Printf("  %-18s config=%d already_in_catalog=%d to_insert=%d\n", c.key, len(raw), len(existing), toAdd)
	}

	if *apply && len(plan) > 0 {
		if *backupPath == "" {
			*backupPath = filepath.Join(os.TempDir(),
				fmt.Sprintf("migrate_device_catalogs_backup_%s_%s.json", dbName, time.Now().Format("20060102_150405")))
		}
		snapshot := bson.M{"configuration": config, "catalog_before": existingByCat}
		data, err := bson.MarshalExtJSON(snapshot, true, false)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(*backupPath, data, 0o644); err != nil {
			panic(err)
		}
		fmt.Println("backup written to", *backupPath)

		now := time.Now().UTC()
		docs := make([]interface{}, 0, len(plan))
		for _, p := range plan {
			docs = append(docs, bson.M{
				"_id": primitive.NewObjectID(), "category": p.cat.key, "value": p.value,
				"value_key": valueKey(p.value), "active": true, "order": p.order, "created_at": now,
			})
		}
		res, err := catalog.InsertMany(ctx, docs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("inserted %d catalog values\n", len(res.InsertedIDs))
	} else if len(plan) == 0 {
		fmt.Println("nothing to insert (already migrated)")
	}

	// ---- report (after the seed when applied) ----
	fmt.Println("\nREPORT (nothing deleted)")
	for _, c := range categories {
		used := map[string]int{}
		cur, err := db.Collection("devices").Aggregate(ctx, mongo.Pipeline{
			{{Key: "$match", Value: bson.M{c.deviceField: bson.M{"$type": "string", "$ne": ""}}}},
			{{Key: "$group", Value: bson.M{"_id": "$" + c.deviceField, "count": bson.M{"$sum": 1}}}},
		})
		if err != nil {
			panic(err)
		}
		var rows []struct {
			Value string `bson:"_id"`
			Count int    `bson:"count"`
		}
		_ = cur.All(ctx, &rows)
		for _, r := range rows {
			used[r.Value] = r.Count
		}

		var values []string
		if *apply {
			var docs []bson.M
			cur, _ := catalog.Find(ctx, bson.M{"category": c.key}, options.Find().SetSort(bson.M{"order": 1}))
			_ = cur.All(ctx, &docs)
			for _, d := range docs {
				values = append(values, fmt.Sprint(d["value"]))
			}
		} else {
			for _, e := range existingByCat[c.key] {
				values = append(values, fmt.Sprint(e["value"]))
			}
			for _, p := range plan {
				if p.cat.key == c.key {
					values = append(values, p.value)
				}
			}
		}

		inCatalog := map[string]bool{}
		var unused, testLike []string
		for _, v := range values {
			inCatalog[v] = true
			if used[v] == 0 {
				unused = append(unused, fmt.Sprintf("%q", v))
				if reason := testLikeReason(v); reason != "" {
					testLike = append(testLike, fmt.Sprintf("%q (%s)", v, reason))
				}
			}
		}
		var missing []string
		for v, n := range used {
			if !inCatalog[v] {
				missing = append(missing, fmt.Sprintf("%q×%d", v, n))
			}
		}
		sort.Strings(missing)

		fmt.Printf("\n[%s] %s: %d values\n", c.key, c.label, len(values))
		fmt.Printf("  in catalog, used by 0 devices (%d): %s\n", len(unused), strings.Join(unused, ", "))
		fmt.Printf("  of those, look like test data (%d): %s\n", len(testLike), strings.Join(testLike, ", "))
		fmt.Printf("  used by devices but NOT in this catalog (%d): %s\n", len(missing), strings.Join(missing, ", "))
	}
}
