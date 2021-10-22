package db_test

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-logr/logr"

	"github.com/gp42/pf-demo/pkg/db"
	"github.com/gp42/pf-demo/pkg/util"
)

var (
	ctx       = context.TODO()
	log       = logr.Discard()
	emptyConn = db.DBConnectionParams{
		Host:     util.StrPtr(""),
		Port:     util.IntPtr(0),
		Password: util.StrPtr(""),
		User:     util.StrPtr(""),
		DBName:   util.StrPtr(""),
		SSLMode:  util.StrPtr(""),
	}
)

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestUpsertBlacklistRecord(t *testing.T) {
	dbPool, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbPool.Close()

	mock.ExpectExec("INSERT INTO blacklists .* UPDATE SET").WithArgs("10.0.0.1", AnyTime{}).WillReturnResult(sqlmock.NewResult(1, 1))

	// Test function
	dbConn := db.NewDBConnection(dbPool, &emptyConn, &log)
	if err = dbConn.UpsertBlacklistRecord(&ctx, "10.0.0.1", time.Now().UTC()); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// Make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestIsIPBlacklisted(t *testing.T) {
	dbPool, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbPool.Close()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)

	mock.ExpectQuery("SELECT id FROM blacklists").WithArgs("10.0.0.1").WillReturnRows(rows)

	// Test function
	dbConn := db.NewDBConnection(dbPool, &emptyConn, &log)
	if _, err = dbConn.IsIPBlacklisted(&ctx, "10.0.0.1"); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// Make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
