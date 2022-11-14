package siid

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const mysqlAddress = "127.0.0.1:3306"

func getMysqlDriver(address string) *mysqlDriver {
	url := fmt.Sprintf("root:@tcp(%s)/mysql?charset=utf8", address)
	if db, err := sql.Open("mysql", url); err != nil {
		panic(err)
	} else {
		if err = db.Ping(); err != nil {
			panic(err)
		}
		driver := NewMysqlDriver(db)
		if err = driver.Prepare(context.Background()); err != nil {
			panic(err)
		}
		return driver.(*mysqlDriver)
	}
}

func Test_MysqlDriverOffset(t *testing.T) {
	driver := getMysqlDriver(mysqlAddress)
	t.Cleanup(func() {
		if err0 := driver.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
	})
	Convey("mysql driver offset", t, func() {
		current, err := driver.Renew(context.Background(), fmt.Sprintf("test_ts_%d", nowFunc().Unix()), 1000, defaultOffsetWhenAutoCreateDomain)
		So(err, ShouldBeNil)
		So(current, ShouldEqual, defaultOffsetWhenAutoCreateDomain)
	})
}

func Test_MysqlDriverShouldLockRow(t *testing.T) {
	driver1 := getMysqlDriver(mysqlAddress)
	driver2 := getMysqlDriver(mysqlAddress)
	driver3 := getMysqlDriver(mysqlAddress)
	driver4 := getMysqlDriver(mysqlAddress)

	t.Cleanup(func() {
		if err0 := driver1.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
		if err0 := driver2.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
		if err0 := driver3.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
		if err0 := driver4.Destroy(context.Background()); err0 != nil {
			t.Error(err0)
		}
	})

	_, _ = driver1.db.Exec(fmt.Sprintf(sqlFmtInsertDomain, defaultName, defaultName, "test1", 0))
	_, _ = driver2.db.Exec(fmt.Sprintf(sqlFmtInsertDomain, defaultName, defaultName, "test2", 0))

	var driver1LockOk bool
	var driver2LockOk bool
	var driver3LockOk bool
	var driver4LockOk bool
	continueChan1 := make(chan struct{})
	continueChan2 := make(chan struct{})
	continueChan3 := make(chan struct{})
	continueChan4 := make(chan struct{})
	driver1.onLockOk = func() {
		driver1LockOk = true
		<-continueChan1
	}
	driver2.onLockOk = func() {
		driver2LockOk = true
		<-continueChan2
	}
	driver3.onLockOk = func() {
		driver3LockOk = true
		<-continueChan3
	}
	driver4.onLockOk = func() {
		driver4LockOk = true
	}

	go func() {
		_, _ = driver1.Renew(context.Background(), "test1", 1, 1)
	}()
	go func() {
		_, _ = driver2.Renew(context.Background(), "test2", 1, 1)
	}()
	go func() {
		_, _ = driver3.Renew(context.Background(), "test3", 1, 1)
	}()
	time.Sleep(time.Duration(100) * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100)*time.Millisecond)
	defer cancel()
	_, err4 := driver4.Renew(ctx, "test2", 1, 1)
	if err4 == nil {
		t.Fatal("err should not nil")
	}
	if !driver1LockOk {
		t.Fatal("driver1 lock failed")
	}
	if !driver2LockOk {
		t.Fatal("driver2 lock failed")
	}
	if !driver3LockOk {
		t.Fatal("driver3 lock failed")
	}
	if driver4LockOk {
		t.Fatal("driver4 should lock failed")
	}
	close(continueChan1)
	close(continueChan2)
	close(continueChan3)
	close(continueChan4)
	time.Sleep(time.Duration(10) * time.Millisecond)
}
