// Database connection logic, data accessor functions and model
package db

import (
	"database/sql"
	"fmt"

	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
)

// Database connection object which keeps datbase connection handler and
// provides data accessor functions
type DBConnection struct {
	param         *DBConnectionParams
	dbPool        *sql.DB
	connectionStr string
	log           *logr.Logger
}

// Parameters for a database connection
type DBConnectionParams struct {
	Host     *string
	Port     *int
	Password *string
	User     *string
	DBName   *string
	SSLMode  *string
}

// Create a new database connection
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

// Initialize connection to the database
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
