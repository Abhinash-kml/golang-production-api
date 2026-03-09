package connections

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresConnection struct {
	DB *sql.DB
}

func (c *PostgresConnection) Connect(connectionString string) error {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	err = db.PingContext(context.Background())
	if err != nil {
		return err
	}

	c.DB = db

	return nil
}

func (c *PostgresConnection) HealthCheck() bool {
	if err := c.DB.PingContext(context.Background()); err != nil {
		return false
	}

	return true
}
