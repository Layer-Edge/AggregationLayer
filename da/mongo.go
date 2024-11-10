package da

import (
	// "encoding/hex"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Layer-Edge/bitcoin-da/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSender struct {
	collection *mongo.Collection
}

func (mongoc *MongoSender) Init(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://vishal:m8C6EhD6ErntgDBu@cluster0.xoean.mongodb.net/")) // "mongodb://foo:bar@localhost:27017"))
	if err != nil {
		return err
	}
	mongoc.collection = client.Database(cfg.Mongo.DB).Collection(cfg.Mongo.Collection)
	return nil
}

func BytesToJson(data []byte) (bson.M, error) {
	var bdoc bson.M
	err := bson.Unmarshal(data, &bdoc)
	if err != nil {
		return bson.M{}, err
	}
	return bdoc, nil
}

func (mongoc *MongoSender) SendData(data []byte) error {
	doc, err := BytesToJson(data)
	if err != nil {
		return err
	}
	res, err := mongoc.collection.InsertOne(context.TODO(), &doc)
	if err != nil {
		log.Panic(err)
		return err
	}
	fmt.Printf("inserted document with ID %v\n", res.InsertedID)
	return nil
}
