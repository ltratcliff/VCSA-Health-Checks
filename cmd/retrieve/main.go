package retrieve

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func GetData() (bson.M, error) {
	var result bson.M
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("devlab").Collection("vcsa")

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}}) // Sort by _id in descending order
	findOptions.SetLimit(1)                  // Limit the number of results to 10

	// Perform the query
	cursor, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(cursor, context.TODO())

	// Iterate through the cursor and print the documents
	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Found document: ", result["_id"])
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Disconnect from MongoDB
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	return result, err
}
