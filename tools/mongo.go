package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSender struct {
	collection *mongo.Collection
}

func (s *MongoSender) Init(uri, database, collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	// Get collection
	s.collection = client.Database(database).Collection(collectionName)
	return nil
}

func (s *MongoSender) SendData(data bson.M) error {
	if s.collection == nil {
		return fmt.Errorf("collection not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.collection.InsertOne(ctx, data)
	return err
}

func main() {
	snd := MongoSender{}

	err := snd.Init("mongodb+srv://admin:admin@localhost:27017/", "devnet", "test")
	if err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	// Parse the transaction result into a proper BSON document
	txResult := bson.M{"success": true, "from": "cosmos1uuhr5kleawdryk3fehunyk8ezav2zzn8d6fj5j", "to": "cosmos1c3y4q50cdyaa5mpfaa2k8rx33ydywl35hsvh0d", "amount": "1 token", "transactionHash": "80DDA4EBA5E74308A7FF711DA5132A64C7E09B5F2F46ACAF8B9B5B8D048350CD", "memo": "D\u0000Hello world!! 2024-12-07 13:43:43.876833982 +0000 UTC m=+5.000642191", "blockHeight": "1142092", "gasUsed": "72571"}

	if err := snd.SendData(txResult); err != nil {
		log.Fatal("Failed to send data:", err)
	}

	log.Println("Data successfully sent to MongoDB")
}
