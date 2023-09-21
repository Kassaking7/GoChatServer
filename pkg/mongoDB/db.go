package mongoDB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient *mongo.Client

func InitDB() error {
	clientOptions := options.Client().ApplyURI("mongodb+srv://kassaking7:196851444@gochatdb.qv4shfe.mongodb.net/?retryWrites=true&w=majority")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

	dbClient = client
	fmt.Println("Connected to MongoDB!")

	return nil
}

func GetMongoClient() *mongo.Client {
	return dbClient
}
