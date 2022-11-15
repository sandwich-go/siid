package siid

import (
	"context"
	"github.com/bwmarrin/snowflake"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"testing"
)

func BenchmarkSIID_MySQL(b *testing.B) {
	initBenchmark()
	bd := New(mysqlDriverName,
		WithDevelopment(false),
		WithEnableMonitor(false),
		WithMaxQuantum(900000),
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
		nowFunc().Nanosecond()
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
