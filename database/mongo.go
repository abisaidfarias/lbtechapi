package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//Conn on mongo
func Conn() *mongo.Database {

	godotenv.Load()
	log.Println("MONGO", os.Getenv("MONGO_URI"))
	clientOpts := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	client, err := mongo.Connect(context.TODO(), clientOpts)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	database := client.Database("lbtechdev")

	indexEmail := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		}, Options: options.Index().SetUnique(true),
	}

	database.Collection("users").Indexes().DropAll(context.Background())
	database.Collection("users").Indexes().CreateOne(context.Background(), indexEmail)

	return database
}
