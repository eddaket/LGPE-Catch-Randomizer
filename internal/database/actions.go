package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	dbName         = "randomizer"
	collectionName = "generations"
)

func (s *service) InsertGeneration(generation *Generation) error {
	db := s.client.Database(dbName)
	collection := db.Collection(collectionName)

	_, err := collection.InsertOne(context.Background(), generation)
	return err
}

func (s *service) GetGenerationById(id string) (*Generation, error) {
	db := s.client.Database(dbName)
	collection := db.Collection(collectionName)

	var generation Generation
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&generation)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &generation, nil
}
