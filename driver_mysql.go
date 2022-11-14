package siid

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sandwich-go/logbus/glog"
	"time"
)

// using innodb row level lock
// https://dev.mysql.com/doc/refman/8.0/en/set-transaction.html
// https://stackoverflow.com/questions/22242081/select-for-update-holding-entire-table-in-mysql-rather-than-row-by-row
// For locking reads (SELECT with FOR UPDATE or LOCK IN SHARE MODE), UPDATE, and DELETE statements, locking depends on whether
// the statement uses a unique index with a unique search condition, or a range-type search condition. For a unique index with a
// unique search condition, InnoDB locks only the index record found, not the gap before it. For other search conditions, InnoDB
// locks the index range scanned, using gap locks or next-key (gap plus index-record) locks to block insertions by other sessions
// into the gaps covered by the range.
const (
	defaultTimeout                   = time.Duration(15) * time.Second
	defaultName                      = "siid"
	sqlCreateMysqlDatabaseIfNotExist = `CREATE DATABASE IF NOT EXISTS %s`
	sqlCreateMysqlTableIfNotExist    = `CREATE TABLE IF NOT EXISTS %s.%s (
	domain varchar(30) NOT NULL,
	id bigint unsigned NOT NULL,
	PRIMARY KEY domain (domain)) ENGINE = Innodb DEFAULT CHARSET = utf8;`
	sqlFmtSelForUp     = "SELECT id FROM %s.%s where domain='%s' FOR UPDATE"
	sqlFmtAddID        = "UPDATE %s.%s SET id = id + %d where domain='%s'"
	sqlFmtInsertDomain = "INSERT INTO %s.%s(domain,id) VALUES('%s',%d)"
)

var emptyCancelFunc = context.CancelFunc(func() {})

type mysqlDriver struct {
	dbName, tableName string
	db                *sql.DB
	onLockOk          func()
}

func NewMysqlDriver(client *sql.DB) Driver {
	return NewMysqlDriverWithName(client, defaultName, defaultName)
}

func NewMysqlDriverWithName(client *sql.DB, dbName, tableName string) Driver {
	return &mysqlDriver{db: client, dbName: dbName, tableName: tableName}
}

func wrapperContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, defaultTimeout)
	}
	return ctx, emptyCancelFunc
}

func (d *mysqlDriver) Prepare(ctx context.Context) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = wrapperContext(ctx)
	if _, err = d.db.ExecContext(ctx, fmt.Sprintf(sqlCreateMysqlDatabaseIfNotExist, d.dbName)); err == nil {
		_, err = d.db.ExecContext(ctx, fmt.Sprintf(sqlCreateMysqlTableIfNotExist, d.dbName, d.tableName))
	}
	cancel()
	return err
}

func (d *mysqlDriver) Destroy(_ context.Context) error { return d.db.Close() }
func (d *mysqlDriver) Renew(ctx context.Context, domain string, quantum, offsetOnCreate uint64) (uint64, error) {
	var cancel context.CancelFunc
	ctx, cancel = wrapperContext(ctx)
	curr, err := d.renew(ctx, domain, quantum)
	if err == errDomainLost {
		_, _ = d.db.Exec(fmt.Sprintf(sqlFmtInsertDomain, d.dbName, d.tableName, domain, offsetOnCreate))
		curr, err = d.renew(ctx, domain, quantum)
	}
	cancel()
	return curr, err
}

func (d *mysqlDriver) renew(ctx context.Context, domain string, quantum uint64) (id uint64, err error) {
	var tx *sql.Tx
	var rows *sql.Rows
	// begin transaction
	if tx, err = d.db.BeginTx(ctx, nil); err != nil {
		return 0, err
	}

	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
		if err != nil {
			if err0 := tx.Rollback(); err0 != nil && err0 != sql.ErrTxDone {
				glog.Error(w("mysql rollback error"), glog.Err(err0))
			}
		}
	}()

	// row lock
	if rows, err = tx.QueryContext(ctx, fmt.Sprintf(sqlFmtSelForUp, d.dbName, d.tableName, domain)); err != nil {
		return 0, err
	}

	if d.onLockOk != nil {
		d.onLockOk()
	}

	found := false
	// must clear query result
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return 0, err
		}
		found = true
	}
	if errScan := rows.Err(); errScan != nil {
		return 0, errScan
	}
	if !found {
		return 0, errDomainLost
	}
	if result, errExec := tx.ExecContext(ctx, fmt.Sprintf(sqlFmtAddID, d.dbName, d.tableName, quantum, domain)); errExec != nil {
		return 0, errExec
	} else {
		if affected, errAffected := result.RowsAffected(); errAffected != nil {
			return 0, errAffected
		} else if affected != 1 {
			return 0, fmt.Errorf("expected to affect 1 row, affected %d", affected)
		}
	}
	if err = tx.Commit(); err != nil {
		return
	}
	return id, nil
}
