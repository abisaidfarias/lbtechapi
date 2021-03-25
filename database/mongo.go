package database

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"

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
	indexCode := mongo.IndexModel{
		Keys: bson.M{
			"code": 1,
		}, Options: options.Index().SetUnique(true),
	}
	indexTestPlanName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true),
	}
	indexTestCategoryName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true),
	}
	indexProfileName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true),
	}
	indexCompanyName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true),
	}
	indexBrandName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true),
	}

	database.Collection("users").Indexes().DropAll(context.Background())
	database.Collection("users").Indexes().CreateOne(context.Background(), indexEmail)
	database.Collection("test_cases").Indexes().DropAll(context.Background())
	database.Collection("test_cases").Indexes().CreateOne(context.Background(), indexCode)
	database.Collection("test_categories").Indexes().DropAll(context.Background())
	database.Collection("test_categories").Indexes().CreateOne(context.Background(), indexTestCategoryName)
	database.Collection("test_plans").Indexes().DropAll(context.Background())
	database.Collection("test_plans").Indexes().CreateOne(context.Background(), indexTestPlanName)
	database.Collection("profiles").Indexes().DropAll(context.Background())
	database.Collection("profiles").Indexes().CreateOne(context.Background(), indexProfileName)
	database.Collection("companies").Indexes().DropAll(context.Background())
	database.Collection("companies").Indexes().CreateOne(context.Background(), indexCompanyName)
	database.Collection("brands").Indexes().DropAll(context.Background())
	database.Collection("brands").Indexes().CreateOne(context.Background(), indexBrandName)
	return database
}
func GetMongoDBClient() *mongo.Client {

	return instance.Client()
}
