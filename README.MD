# SIID
Fast, thread-safe Integer ID generator in Golang, which is globally unique and non-strictly incremental (in the case of multiple concurrent processes).

## Features
- Thread safe
- Generate globally unique integer ID, no duplicate ID
- IDs are non-strictly incremental in a multi-processes environment
- IDs do not interfere with each other on different `domains`
- No time dependency, no clock redirection, no ID rewinding
- Built-in `MySQL` and `Mongo` drivers
- Implement `Driver` interface, you can implement the new driver
- Automatic expansion and contraction of ID segments according to the frequency of ID generation, maintain high performance when generation is frequent
- `MaxQuantum` to avoid wasted segments caused by unexpected crashes
- Generate a continuous segment of IDs in memory to ensure high performance
- When the ID reaches the percentage of `RenewPercent`, fork new goroutine to get the next ID segment from the driver to avoid business jams
- Monitoring `Renew` errors、the number of `Renew` cost or calls、the number of ID generation cost or calls、the current ID segment、the current ID maximum, and the number of remaining IDs

## Links
* [English](https://github.com/sandwich-go/siid/blob/master/README.MD)
* [中文文档](https://github.com/sandwich-go/siid/blob/master/README_CN.MD)

## Benchmark
```shell
Benchmark Results

goos: darwin
goarch: amd64
pkg: github.com/sandwich-go/siid
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkSIID_MySQL
BenchmarkSIID_MySQL-16    	16460401	        63.43 ns/op	       0 B/op	       0 allocs/op
BenchmarkRand-16          	92374222	        12.26 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimestamp-16     	17644713	        65.97 ns/op	       0 B/op	       0 allocs/op
BenchmarkUUID_V1-16       	12334704	        97.00 ns/op	       0 B/op	       0 allocs/op
BenchmarkUUID_V2-16       	12190599	        97.27 ns/op	       0 B/op	       0 allocs/op
BenchmarkUUID_V3-16       	 5514180	       216.7 ns/op	     144 B/op	       4 allocs/op
BenchmarkUUID_V4-16       	 1595637	       729.5 ns/op	      16 B/op	       1 allocs/op
BenchmarkUUID_V5-16       	 4900987	       249.4 ns/op	     168 B/op	       4 allocs/op
BenchmarkSnowflake-16     	 4931772	       244.1 ns/op	       0 B/op	       0 allocs/op
PASS
```
- `SIID` is a globally unique integer ID generation solution and is not a replacement for `UUID`
- `SIID` are about 4 times faster than those of [UUUID V5](https://github.com/satori/go.uuid) and about 4 times faster than those of the [Snowflake](https://github.com/bwmarrin/snowflake) solution.

## UseCase
```go
import (
    "context"
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "github.com/sandwich-go/siid"
)

func main()  {
    const driverName = "mysql"
    // register driver
    dsn := fmt.Sprintf("root:@tcp(%s)/mysql?charset=utf8", "127.0.0.1:3306")
    if db, err0 := sql.Open("mysql", dsn); err0 != nil {
        panic(err0)
    } else {
        driver := siid.NewMysqlDriver(db)
        siid.Register(driverName, driver)	
    }   
    // create builder
    b := siid.New(driverName,
        WithOffsetWhenAutoCreateDomain(30000000),
        WithDevelopment(false),
    )
    err1 := b.Prepare(context.Background())
    if err1 != nil {
        panic(err1)	
    }
    defer func(){
        _ = b.Destroy(context.Background())
    }() 
    // use engine, generate id for `domain`
    var domain = "player"
    e, err2 := b.Build(domain)
    if err2 != nil {
        panic(err2)
    }
    id, err3 := e.Next()
    if err3 != nil {
        panic(err3)
    }
    fmt.Println("id =>", id)

    id1, err4 := e.Next()
    if err4 != nil {
        panic(err4)
    }
    fmt.Println("id =>", id1)
}
```