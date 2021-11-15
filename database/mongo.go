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
	collation := options.Collation{
		Strength:        2,
		Locale:          "es",
		CaseLevel:       false,
		CaseFirst:       "off",
		NumericOrdering: false,
		Alternate:       "non-ignorable",
		MaxVariable:     "punct",
		Normalization:   false,
		Backwards:       false,
	}

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	database := client.Database(os.Getenv("MONGO_DB"))

	indexEmail := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexCode := mongo.IndexModel{
		Keys: bson.M{
			"code": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexTestPlanName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexTestCategoryName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexProfileName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexCompanyName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexBrandName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexCountryName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexDeviceCommercialName := mongo.IndexModel{
		Keys: bson.M{
			"commercial_model": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexPrinterSerial := mongo.IndexModel{
		Keys: bson.M{
			"serial": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	indexLocationName := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}

	indexImei := mongo.IndexModel{
		Keys: bson.M{
			"imei": 1,
		}, Options: options.Index().SetUnique(true).SetCollation(&collation),
	}
	database.Collection("users").Indexes().CreateOne(context.Background(), indexEmail)
	database.Collection("test_cases").Indexes().CreateOne(context.Background(), indexCode)
	database.Collection("test_categories").Indexes().CreateOne(context.Background(), indexTestCategoryName)
	database.Collection("test_plans").Indexes().CreateOne(context.Background(), indexTestPlanName)
	database.Collection("profiles").Indexes().CreateOne(context.Background(), indexProfileName)
	database.Collection("companies").Indexes().CreateOne(context.Background(), indexCompanyName)
	database.Collection("brands").Indexes().CreateOne(context.Background(), indexBrandName)
	database.Collection("countries").Indexes().CreateOne(context.Background(), indexCountryName)
	database.Collection("devices").Indexes().CreateOne(context.Background(), indexDeviceCommercialName)
	database.Collection("printers").Indexes().CreateOne(context.Background(), indexPrinterSerial)
	database.Collection("locations").Indexes().CreateOne(context.Background(), indexLocationName)
	database.Collection("device_trackings").Indexes().CreateOne(context.Background(), indexImei)
	return database
}
func GetMongoDBClient() *mongo.Client {

	return instance.Client()
}
