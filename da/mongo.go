package da

import (
	// "encoding/hex"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// type MongoSender struct {
// 	collection *mongo.Collection
// }

type MongoSender struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoSender(uri, dbName, collectionName string) (*MongoSender, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Select database and collection
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	return &MongoSender{
		client:     client,
		collection: collection,
	}, nil
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
	// Check if collection is initialized
	if mongoc.collection == nil {
		return fmt.Errorf("MongoDB collection not initialized")
	}

	// Parse the JSON data into a map
	var doc map[string]interface{}
	if err := bson.UnmarshalExtJSON(data, true, &doc); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Add timestamp field
	doc["timestamp"] = time.Now().UTC() // Adds current UTC time
	// Alternative: Use ISODate format that MongoDB prefers
	// doc["timestamp"] = primitive.NewDateTimeFromTime(time.Now().UTC())

	// Insert the document
	res, err := mongoc.collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return fmt.Errorf("failed to insert document: %v", err)
	}

	fmt.Printf("inserted document with ID %v\n", res.InsertedID)
	return nil
}
