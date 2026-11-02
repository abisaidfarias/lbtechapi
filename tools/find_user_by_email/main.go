package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	_ = godotenv.Load()
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	email := "fariasabisaid@gmail.com"
	deleteID := ""
	if len(os.Args) > 1 {
		email = os.Args[1]
	}
	if len(os.Args) > 2 {
		deleteID = os.Args[2]
	}
	email = strings.ToLower(strings.TrimSpace(email))
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	coll := client.Database(dbName).Collection("users")
	if deleteID != "" {
		oid, err := primitive.ObjectIDFromHex(deleteID)
		if err != nil {
			panic(err)
		}
		res, err := coll.DeleteOne(ctx, bson.M{"_id": oid})
		if err != nil {
			panic(err)
		}
		fmt.Printf("deleted %d document(s) _id=%s\n", res.DeletedCount, deleteID)
		return
	}
	collation := options.Collation{Locale: "es", Strength: 2}
	cur, err := coll.Find(ctx, bson.M{"email": email}, options.Find().SetCollation(&collation))
	if err != nil {
		panic(err)
	}
	var docs []bson.M
	_ = cur.All(ctx, &docs)
	fmt.Printf("found %d user(s) for %q\n", len(docs), email)
	for _, d := range docs {
		fmt.Printf("  _id=%v email=%v account_status=%v\n", d["_id"], d["email"], d["account_status"])
	}
}
