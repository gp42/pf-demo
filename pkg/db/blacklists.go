package db

import (
	"context"
	"database/sql"
	"time"
)

// UpsertBlacklistRecord adds or updates a Blacklisted record in the database
func (d *DBConnection) UpsertBlacklistRecord(ctx *context.Context, blockIP string, t time.Time) error {
	sqlStatement := `
		INSERT INTO blacklists (block_ip, block_timestamp)
		VALUES ($1, $2)
		ON CONFLICT (block_ip)
		DO 
			UPDATE SET block_timestamp = $2;
	`
	_, err := d.dbPool.ExecContext(*ctx, sqlStatement, blockIP, t)
	if err != nil {
		return err
	}

	return nil
}

// IsIPBlacklisted checks if the provided IP address is present in the blacklisted database
func (d *DBConnection) IsIPBlacklisted(ctx *context.Context, blockIP string) (bool, error) {
	sqlStatement := `
		SELECT id FROM blacklists
		WHERE block_ip = $1
	`

	row := d.dbPool.QueryRowContext(*ctx, sqlStatement, blockIP)
	var id int
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
