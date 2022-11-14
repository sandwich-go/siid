//go:build mongo
// +build mongo

package siid

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

const mongoAddress = "127.0.0.1:32797"

func getMongoClient(addr string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	uri := fmt.Sprintf("mongodb://%s", addr)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return client
}

func getMongolDriver(address string) *mysqlDriver {
	driver := NewMongoDriver(getMongoClient(address))
	if err := driver.Prepare(context.Background()); err != nil {
		panic(err)
	}
	return driver.(*mysqlDriver)
}

func Test_MongoDriverOffset(t *testing.T) {
	driver := getMongolDriver(mongoAddress)
	t.Cleanup(func() {
		if err0 := driver.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
	})
	Convey("mongo driver offset", t, func() {
		current, err := driver.Renew(context.Background(), fmt.Sprintf("test_ts_%d", nowFunc().Unix()), 1000, defaultOffsetWhenAutoCreateDomain)
		So(err, ShouldBeNil)
		So(current, ShouldEqual, defaultOffsetWhenAutoCreateDomain)
	})
}
