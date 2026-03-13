package connections

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgresConnection struct {
	connectionString string
	DB               *sql.DB
}

func NewPostgresConnection(connectionString string) *PostgresConnection {
	connection := &PostgresConnection{connectionString: connectionString}
	err := connection.Connect()
	if err != nil {
		zap.L().Fatal("Postgres connection failed", zap.Error(err))
		return nil
	}

	return connection
}

func (c *PostgresConnection) Connect() error {
	db, err := sql.Open("postgres", c.connectionString)
	if err != nil {
		return err
	}

	err = db.PingContext(context.Background())
	if err != nil {
		return err
	}

	c.DB = db

	c.onConnnect()

	return nil
}

func (c *PostgresConnection) onConnnect() {
	zap.L().Info("Connected to postgresql database", zap.String("dsn", c.connectionString))
}

func (c *PostgresConnection) HealthCheck() bool {
	if err := c.DB.PingContext(context.Background()); err != nil {
		return false
	}

	return true
}
