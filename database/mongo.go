package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//Conn on mongo
func Conn() *mongo.Database {

	clientOpts := options.Client().ApplyURI("mongodb://lbtechbd-dev:cZ6JkSeY6qtSNO1KRhFqR3RksRtMbQjPXYc0jhDM9RRMXH1i1RxHpOmmQxi3YeRRYQmWL42TK6rr5xIWSKwoyg%3D%3D@lbtechbd-dev.mongo.cosmos.azure.com:10255/?ssl=true&retrywrites=false&maxIdleTimeMS=120000&appName=@lbtechbd-dev@")
	client, err := mongo.Connect(context.TODO(), clientOpts)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	database := client.Database("lbtechdev")

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
			"dni":   1,
		}, Options: options.Index().SetUnique(true),
	}

	database.Collection("users").Indexes().CreateOne(context.Background(), indexModel)

	return database
}
