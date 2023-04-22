package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type MongoClient struct {
	client          *mongo.Client
	worksCollection *mongo.Collection
}

func NewMongoClient(ctx context.Context) (c *MongoClient, err error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		err = fmt.Errorf("empty MONGO_URI for connection string")
		return
	}
	databaseName := os.Getenv("MONGO_DATABASE")
	if databaseName == "" {
		err = fmt.Errorf("empty MONGO_DATABASE for connection string")
		return
	}
	collectionName := os.Getenv("MONGO_WORKS_COLLECTION")
	if collectionName == "" {
		err = fmt.Errorf("empty MONGO_WORKS_COLLECTION for connection string")
		return
	}

	opts := options.Client().ApplyURI(uri).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	retryWrites := os.Getenv("MONGO_RETRY_WRITES")
	if retries, err := strconv.ParseBool(retryWrites); err == nil && retries {
		opts.SetRetryWrites(retries)
	}
	timeoutMs := os.Getenv("MONGO_TIMEOUT_MS")
	if timeout, err := strconv.ParseInt(timeoutMs, 10, 32); err == nil && timeout != 0 {
		opts.SetTimeout(time.Duration(timeout) * time.Millisecond)
	}

	c = &MongoClient{}

	c.client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return
	}

	c.worksCollection = c.client.Database(databaseName).Collection(collectionName)
	return
}

var _ repository.ReadWriteRepository = (*MongoClient)(nil)

func (m *MongoClient) Add(ctx context.Context, work *models.WorkItem) (result *models.WorkItem, err error) {

	result = work
	out, err := m.worksCollection.InsertOne(ctx, work)
	if err != nil {
		return
	}
	id, ok := out.InsertedID.(primitive.ObjectID)
	if !ok {
		return
	}
	log.Printf("successfully inserted work with document id %v\n", id)
	return
}

func (m *MongoClient) Update(ctx context.Context, work *models.WorkItem) (result *models.WorkItem, err error) {
	filter := bson.D{{Key: "_id", Value: work.Id}}
	update := bson.M{
		"$set": work,
	}
	opts := options.Update().SetUpsert(true)
	out, err := m.worksCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return
	}
	result = work
	log.Printf("successfully updated %v work document with id %v\n", out.ModifiedCount, work.Id)
	return
}

func (m *MongoClient) GetById(ctx context.Context, id string) (results []*models.WorkItem, err error) {
	filter := bson.D{{Key: "workId", Value: id}}
	cursor, err := m.worksCollection.Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var wi models.WorkItem
		if err = cursor.Decode(&wi); err != nil {
			return
		}
		results = append(results, &wi)
	}
	if err = cursor.Err(); err != nil {
		return
	}
	return
}

func (m *MongoClient) List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) (result []*models.WorkItem, err error) {
	orderedFilter := bson.A{
		bson.D{{Key: "startDate", Value: bson.D{{Key: "$gte", Value: from}}}},
		bson.D{{Key: "startDate", Value: bson.D{{Key: "$lte", Value: to}}}},
	}
	if len(zones) > 0 {
		orderedFilter = append(orderedFilter, bson.D{{Key: "zones", Value: bson.D{{Key: "$in", Value: zones}}}})
	}
	if len(statuses) > 0 {
		orderedFilter = append(orderedFilter, bson.D{{Key: "status", Value: bson.D{{Key: "$in", Value: statuses}}}})
	}
	filter := bson.D{{Key: "$and", Value: orderedFilter}}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "startDate", Value: 1}})

	log.Printf("searching for work documents with filer %+v\n", filter)
	cursor, err := m.worksCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var wi models.WorkItem
		if err = cursor.Decode(&wi); err != nil {
			return
		}
		result = append(result, &wi)
	}
	if err = cursor.Err(); err != nil {
		return
	}
	return
}
