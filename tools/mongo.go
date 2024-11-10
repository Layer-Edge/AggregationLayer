package main

import (
    // "encoding/hex"
    "fmt"
    "log"
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "time"
)

type MongoSender struct {
	collection *mongo.Collection
}

func (mongoc *MongoSender) Init(Endpoint string, DB string, Collection string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(Endpoint)) // "mongodb://foo:bar@localhost:27017"))
	if err != nil { return err }
	mongoc.collection = client.Database(DB).Collection(Collection)
	return nil
}

func (mongoc *MongoSender) SendData(doc bson.M) error {
	res, err := mongoc.collection.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("inserted document with ID %v\n", res.InsertedID)
	return nil
}

func main() {
	snd := MongoSender{collection:nil};
	err := snd.Init("mongodb+srv://vishal:m8C6EhD6ErntgDBu@cluster0.xoean.mongodb.net", "devnet", "test")
	log.Println(err)
	snd.SendData(bson.M{"hello": "world"})
}
