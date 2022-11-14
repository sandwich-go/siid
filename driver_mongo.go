package siid

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// upsert can not with different values for update and insert
// https://jira.mongodb.org/browse/SERVER-453
// https://jira.mongodb.org/browse/SERVER-991
// https://docs.mongodb.com/manual/reference/operator/update/setOnInsert/#up._S_setOnInsert
// setOnInsert cannot perform on same filed
type mongoDriver struct {
	dbName         string
	collectionName string
	client         *mongo.Client
	copts          *options.CollectionOptions
}

func NewMongoDriver(client *mongo.Client) Driver {
	return NewMongoDriverWithName(client, defaultName, defaultName)
}

func NewMongoDriverWithName(client *mongo.Client, dbName, collectionName string) Driver {
	return &mongoDriver{client: client, dbName: dbName, collectionName: collectionName}
}

func (m *mongoDriver) Prepare(ctx context.Context) error { return m.pingPrimary(ctx) }
func (m *mongoDriver) pingPrimary(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = wrapperContext(ctx)
	err := m.client.Ping(ctx, readpref.Primary())
	cancel()
	return err
}

func (m *mongoDriver) Destroy(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = wrapperContext(ctx)
	err := m.client.Disconnect(ctx)
	cancel()
	return err
}

func (m *mongoDriver) getCollection() *mongo.Collection {
	return m.client.Database(m.dbName).Collection(m.collectionName, m.copts)
}

func (m *mongoDriver) Renew(ctx context.Context, domain string, quantum, offset uint64) (uint64, error) {
	var cancel context.CancelFunc
	ctx, cancel = wrapperContext(ctx)
	curr, err := m.renew(ctx, domain, quantum)
	if err == errDomainLost { // create new domain
		// do not care fail
		_, _ = m.getCollection().InsertOne(ctx, bson.D{{Key: "_id", Value: domain}, {Key: "current", Value: offset}})
		curr, err = m.renew(ctx, domain, quantum)
	}
	cancel()
	return curr, err
}

func (m *mongoDriver) renew(ctx context.Context, domain string, quantum uint64) (uint64, error) {
	filter := bson.D{{Key: "_id", Value: domain}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "current", Value: quantum}}}}
	var opts options.FindOneAndUpdateOptions
	opts.SetUpsert(false).SetReturnDocument(options.Before)

	var doc struct{ Current uint64 }
	err := m.getCollection().FindOneAndUpdate(ctx, filter, update, &opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, errDomainLost
		}
		return 0, err
	}
	return doc.Current, nil
}
