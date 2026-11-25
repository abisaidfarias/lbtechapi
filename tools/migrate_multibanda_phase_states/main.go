// migrate_multibanda_phase_states prepares multibandas for SMR without
// testing (Sol 5):
//
//	A. test_start_date, test_end_date, under_start_date, under_end_date and
//	   completed_date stored as Go's zero time (year 1) become null.
//	B. phase_states is filled where missing, from current_phase:
//	   phase < current → done, == current → pending, > current → pending,
//	   and complete = done when current_phase is COMPLETE (4).
//
// Idempotent: A only matches dates before 1971 and B only documents without
// phase_states, so a rerun is a no-op. requires_testing and smr_validation
// are not added (missing = null).
//
// Usage (from the repo root, which holds .env with MONGO_URI / MONGO_DB):
//
//	go run ./tools/migrate_multibanda_phase_states            # dry run: counts only
//	go run ./tools/migrate_multibanda_phase_states -apply     # backup + migrate
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

const collectionName = "multibandas"

var dateFields = []string{"test_start_date", "test_end_date", "under_start_date", "under_end_date", "completed_date"}

// zeroCutoff: Go's zero time is year 1; real dates are far after 1971.
var zeroCutoff = time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)

func zeroDateFilter(field string) bson.M {
	return bson.M{field: bson.M{"$type": "date", "$lt": zeroCutoff}}
}

func missingPhaseStatesFilter() bson.M {
	return bson.M{"phase_states": bson.M{"$exists": false}}
}

var phaseNames = []string{"planning", "sample_reception", "testing", "under_evaluation", "complete"}

const completePhase = 4

// phaseStatesExpr computes phase_states from $current_phase in the update
// pipeline (same rule as functions.ComputePhaseStates).
func phaseStatesExpr() bson.M {
	out := bson.M{}
	for i, name := range phaseNames {
		done := bson.M{"$lt": bson.A{i, "$current_phase"}}
		if i == completePhase {
			done = bson.M{"$gte": bson.A{"$current_phase", completePhase}}
		}
		out[name] = bson.M{"$cond": bson.A{done, "done", "pending"}}
	}
	return out
}

// PhaseStatesFor is the Go version of phaseStatesExpr, used in the preview
// and the tests.
func PhaseStatesFor(currentPhase int) map[string]string {
	out := map[string]string{}
	for i, name := range phaseNames {
		state := "pending"
		if i < currentPhase || (i == completePhase && currentPhase >= completePhase) {
			state = "done"
		}
		out[name] = state
	}
	return out
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
	fmt.Printf("migrate_multibanda_phase_states on database %q: %s\n", dbName, mode)

	total, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("  multibandas=%d\n", total)
	anyZero := bson.A{}
	pending := 0
	for _, f := range dateFields {
		n, err := coll.CountDocuments(ctx, zeroDateFilter(f))
		if err != nil {
			panic(err)
		}
		pending += int(n)
		anyZero = append(anyZero, zeroDateFilter(f))
		fmt.Printf("  A. %-17s zero → null: %d\n", f, n)
	}
	missing, err := coll.CountDocuments(ctx, missingPhaseStatesFilter())
	if err != nil {
		panic(err)
	}
	fmt.Printf("  B. phase_states missing: %d\n", missing)
	pending += int(missing)

	if !*apply || pending == 0 {
		if pending == 0 {
			fmt.Println("nothing to migrate (already migrated)")
		}
		return
	}

	// Backup every affected document's fields before touching anything.
	affected := bson.M{"$or": append(anyZero, missingPhaseStatesFilter())}
	if *backupPath == "" {
		*backupPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("migrate_multibanda_phase_states_backup_%s_%s.jsonl", dbName, time.Now().Format("20060102_150405")))
	}
	f, err := os.Create(*backupPath)
	if err != nil {
		panic(err)
	}
	projection := bson.M{"current_phase": 1, "status": 1, "phase_states": 1}
	for _, field := range dateFields {
		projection[field] = 1
	}
	cur, err := coll.Find(ctx, affected, options.Find().SetProjection(projection))
	if err != nil {
		panic(err)
	}
	backedUp := 0
	for cur.Next(ctx) {
		line, err := bson.MarshalExtJSON(cur.Current, true, false)
		if err != nil {
			panic(err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			panic(err)
		}
		backedUp++
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
	cur.Close(ctx)
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("backup of %d documents written to %s\n", backedUp, *backupPath)

	for _, field := range dateFields {
		res, err := coll.UpdateMany(ctx, zeroDateFilter(field), bson.M{"$set": bson.M{field: nil}})
		if err != nil {
			panic(err)
		}
		fmt.Printf("  A. %-17s matched=%d modified=%d\n", field, res.MatchedCount, res.ModifiedCount)
	}
	res, err := coll.UpdateMany(ctx, missingPhaseStatesFilter(),
		mongo.Pipeline{{{Key: "$set", Value: bson.M{"phase_states": phaseStatesExpr()}}}})
	if err != nil {
		panic(err)
	}
	fmt.Printf("  B. phase_states    matched=%d modified=%d\n", res.MatchedCount, res.ModifiedCount)

	left, _ := coll.CountDocuments(ctx, bson.M{"$or": append(anyZero, missingPhaseStatesFilter())})
	fmt.Printf("remaining to migrate: %d\n", left)
	if left > 0 {
		os.Exit(1)
	}
	fmt.Println("done")
}
