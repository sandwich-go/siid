# SIID
协程安全的整型ID生成器，该ID是全局唯一的，并且是非严格递增的（多协程情况下）。

## 特性
- 协程安全
- 能够生成全局唯一的整型ID，不会出现重复的ID
- 在多协程环境中，ID是非严格递增的
- 不同的`domain`，ID互不干扰
- 不依赖时间，无时钟回拨问题，无ID回绕问题
- 内置`MySQL`、`Mongo`驱动
- 实现`Driver`定义的接口，可自定义驱动
- 根据ID生成的频率，自动扩缩ID段，当ID生成频繁时，仍然保持高性能
- 通过`MaxQuantum`参数避免服务意外崩溃导致的号段浪费
- 每次根据ID段生成一段连续的ID置于内存中，来保证高性能
- 当ID达到`RenewPercent`百分比时，会启动新协程从驱动中获取新的ID段，来避免造成业务卡顿

## 性能测试
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
- `SIID`是全局唯一的数字ID生成方案，并不是`UUID`的替代品
- `SIID`整体性能数据较`UUID V5`要快4倍左右,较`Snowflake`方案快4倍左右。

## 使用方式
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
    // 注册driver
    dsn := fmt.Sprintf("root:@tcp(%s)/mysql?charset=utf8", "127.0.0.1:3306")
    if db, err0 := sql.Open("mysql", dsn); err0 != nil {
        panic(err0)
    } else {
        driver := siid.NewMysqlDriver(db)
        siid.Register(driverName, driver)	
    }   
    // 创建builder
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
    // 通过domain对应的engine来生成domain的ID
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