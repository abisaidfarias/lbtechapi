package database

import (
	"context"
	"log"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//Conn on mongo
func Conn() *mongo.Database {
	utils.LoadConfig()

	uri := viper.GetString("uri")

	clientOpts := options.Client().ApplyURI(uri)

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
			"dni":   1,
		}, Options: options.Index().SetUnique(true),
	}

	indexDni := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
			"dni":   1,
		}, Options: options.Index().SetUnique(true),
	}

	database.Collection("users").Indexes().CreateOne(context.Background(), indexEmail)
	database.Collection("users").Indexes().CreateOne(context.Background(), indexDni)

	return database
}
