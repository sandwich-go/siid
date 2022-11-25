package bench

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bwmarrin/snowflake"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sandwich-go/siid"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	once                          sync.Once
	mysqlDriverName, mysqlAddress = "mysql", "127.0.0.1:3306"
)

func initBenchmark() {
	once.Do(func() {
		siid.Register(mysqlDriverName, getMysqlDriver(mysqlAddress))
	})
}

func getMysqlDriver(address string) siid.Driver {
	url := fmt.Sprintf("root:@tcp(%s)/mysql?charset=utf8", address)
	if db, err := sql.Open("mysql", url); err != nil {
		panic(err)
	} else {
		if err = db.Ping(); err != nil {
			panic(err)
		}
		driver := siid.NewMysqlDriver(db)
		if err = driver.Prepare(context.Background()); err != nil {
			panic(err)
		}
		return driver
	}
}

func BenchmarkSIID_MySQL(b *testing.B) {
	initBenchmark()
	bd := siid.New(mysqlDriverName, siid.NewConfig(
		siid.WithDevelopment(false),
		siid.WithEnableMonitor(false),
		siid.WithMaxQuantum(900000)),
	)
	if err := bd.Prepare(context.Background()); err != nil {
		b.Fatal(err)
	}
	e, err0 := bd.Build("test2")
	if err0 != nil {
		b.Fatal(err0)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = e.Next()
	}
}

func BenchmarkRand(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rand.Int63()
	}
}

func BenchmarkTimestamp(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		time.Now().Nanosecond()
	}
}

func BenchmarkUUID_V1(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uuid.NewV1()
	}
}

func BenchmarkUUID_V2(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uuid.NewV2(128)
	}
}

func BenchmarkUUID_V3(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uuid.NewV3(uuid.NamespaceDNS, "example.com")
	}
}

func BenchmarkUUID_V4(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uuid.NewV4()
	}
}

func BenchmarkUUID_V5(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uuid.NewV5(uuid.NamespaceDNS, "example.com")
	}
}

func BenchmarkSnowflake(b *testing.B) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node.Generate()
	}
}
