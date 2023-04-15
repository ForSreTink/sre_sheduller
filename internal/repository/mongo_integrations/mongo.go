package mongo

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"workScheduler/internal/api/models"
	"workScheduler/internal/repository"

	"github.com/google/uuid"
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
	if uri == "" {
		err = fmt.Errorf("empty MONGO_DATABASE for connection string")
		return
	}
	collectionName := os.Getenv("MONGO_WORKS_COLLECTION")
	if uri == "" {
		err = fmt.Errorf("empty MONGO_WORKS_COLLECTION for connection string")
		return
	}

	opts := options.Client().
		SetWriteConcern(writeconcern.New(writeconcern.WMajority())).
		ApplyURI(uri)
	retryWrites := os.Getenv("MONGO_RETRY_WRITES")
	if retries, err := strconv.ParseBool(retryWrites); err == nil && retries {
		opts.SetRetryWrites(retries)
	}
	timeoutMs := os.Getenv("MONGO_TIMEOUT_MS")
	if timeout, err := strconv.ParseInt(timeoutMs, 10, 32); err == nil && timeout != 0 {
		opts.SetTimeout(time.Duration(timeout) * time.Millisecond)
	}

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
	result.Id = uuid.New().String()
	out, err := m.worksCollection.InsertOne(ctx, work)
	if err != nil {
		return
	}
	id, ok := out.InsertedID.(primitive.ObjectID)
	if !ok {
		return
	}
	fmt.Printf("successfully inserted work with document id %v\n", id)
	return
}

func (m *MongoClient) Update(ctx context.Context, work *models.WorkItem) (result *models.WorkItem, err error) {
	filter := bson.D{{Key: "workId", Value: work.Id}}
	opts := options.Update().SetUpsert(true)
	out, err := m.worksCollection.UpdateOne(ctx, filter, work, opts)
	if err != nil {
		return
	}
	fmt.Printf("successfully updated %v work document with id %v\n", out.ModifiedCount, work.Id)
	return
}

func (m *MongoClient) GetById(ctx context.Context, id string) (result *models.WorkItem, err error) {
	filter := bson.D{{Key: "workId", Value: id}}
	err = m.worksCollection.FindOne(ctx, filter).Decode(&result)
	return
}

func (m *MongoClient) List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) (result []*models.WorkItem, err error) {

	orderedFilter := bson.A{
		bson.D{{Key: "startDate", Value: bson.D{{Key: "$gte", Value: from}}}},
		bson.D{{Key: "startDate", Value: bson.D{{Key: "$lte", Value: from}}}},
	}
	if len(zones) > 0 {
		orderedFilter = append(orderedFilter, bson.E{Key: "zone", Value: bson.M{"$in": zones}})
	}
	if len(statuses) > 0 {
		orderedFilter = append(orderedFilter, bson.E{Key: "status", Value: bson.M{"$in": statuses}})
	}
	filter := bson.D{{Key: "$and", Value: orderedFilter}}

	fmt.Printf("searching for work documents with filer %+v\n", filter)
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
		result = append(result, &wi)
	}
	if err = cursor.Err(); err != nil {
		return
	}
	return
}
