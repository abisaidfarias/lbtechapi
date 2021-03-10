package database

import (
	"github.com/joho/godotenv"
	"os"
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var lock = &sync.Mutex{}

var instance *mongo.Database

// GetInstance returns the unique database instance
func GetInstance() *mongo.Database {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = initDB()
		}
	}

	return instance
}

//Conn on mongo
func initDB() *mongo.Database {

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
