package database

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Close(ctx context.Context) error

	GetGenerationById(id string) (*Generation, error)
	InsertGeneration(generation *Generation) error
}

type service struct {
	client *mongo.Client
}

func New() Service {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Printf("[WARN] Database URI not set, using localhost")
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("[ERROR} Could not connect to database:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("[ERROR} Could not ping database:", err)
	}

	log.Printf("[INFO] Database connection established")
	return &service{client: client}
}

func (s *service) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
