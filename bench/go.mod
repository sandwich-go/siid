module github.com/sandwich-go/siid/bench

go 1.16

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/sandwich-go/siid v0.1.0-alpha.8
	github.com/satori/go.uuid v1.2.0
)

replace github.com/sandwich-go/siid => ./..
