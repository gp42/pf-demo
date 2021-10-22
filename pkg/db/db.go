// Package db contains db connection logic, data accessor functions and model
package db

import (
	"database/sql"
	"fmt"

	"github.com/go-logr/logr"
	// Load postgres driver
	_ "github.com/lib/pq"
)

// DBConnection object which keeps database connection handler and
// provides data accessor functions
type DBConnection struct {
	param         *DBConnectionParams
	dbPool        *sql.DB
	connectionStr string
	log           *logr.Logger
}

// DBConnectionParams are parameters for a database connection
type DBConnectionParams struct {
	Host     *string
	Port     *int
	Password *string
	User     *string
	DBName   *string
	SSLMode  *string
}

// NewDBConnection creates a new database connection
func NewDBConnection(dbPool *sql.DB, c *DBConnectionParams, logger *logr.Logger) *DBConnection {
	db := &DBConnection{
		param:  c,
		dbPool: dbPool,
		connectionStr: fmt.Sprintf(
			"host=%s port=%d user=%s "+
				"password=%s dbname=%s sslmode=%s",
			*c.Host, *c.Port, *c.User, *c.Password, *c.DBName, *c.SSLMode),
		log: logger,
	}
	return db
}

// InitConnection initializes a connection to the database
func (d *DBConnection) InitConnection() error {
	var err error
	d.dbPool, err = sql.Open("postgres", d.connectionStr)
	if err != nil {
		return fmt.Errorf("Failed to open sql connection: %s", err)
	}

	err = d.dbPool.Ping()
	if err != nil {
		return fmt.Errorf("Failed to ping sql db: %s", err)
	}

	d.log.Info("Successfully connected to postgres db.")
	return nil
}

// Close database connection
func (d *DBConnection) Close() {
	d.dbPool.Close()
}
