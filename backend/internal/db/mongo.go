package db

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	dbName string
)

func InitMongo() (*mongo.Database, error) {
	uri := strings.TrimSpace(os.Getenv("MONGO_URI"))
	dbName = strings.TrimSpace(os.Getenv("MONGO_DB"))
	if uri == "" || dbName == "" {
		return nil, errors.New("missing MongoDB configuration (MONGO_URI, MONGO_DB)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := c.Ping(ctx, nil); err != nil {
		return nil, err
	}

	client = c
	return client.Database(dbName), nil
}

func DB() *mongo.Database {
	if client == nil {
		return nil
	}
	return client.Database(dbName)
}
