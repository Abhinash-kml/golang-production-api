package repository

import (
	"database/sql"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	_ "github.com/lib/pq"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type PostgresUserRepository struct {
	db     *sql.DB
	tracer oteltracer.Tracer
}

func NewPostgresUserRepository(connection *connections.PostgresConnection, tracer oteltracer.Tracer) *PostgresUserRepository {
	return &PostgresUserRepository{db: connection.DB, tracer: tracer}
}

func (r *PostgresUserRepository) Setup() error {
	return nil
}

func (r *PostgresUserRepository) GetUsers() ([]model.User, error) {
	query := `SELECT * FROM users;`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Info("Error querying rows")
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	var user model.User

	for rows.Next() {
		rows.Scan(&user.Id, &user.Name, &user.City, &user.State, &user.Country)
		users = append(users, user)
	}

	return users, nil
}

func (r *PostgresUserRepository) GetById(id int) (*model.User, error) {
	query := `SELECT * FROM users WHERE id = $1;`
	var user model.User
	if err := r.db.QueryRow(query, id).Scan(&user.Id, &user.Name, &user.City, &user.State, &user.Country); err != nil {
		return nil, err
	}

	return &user, nil

}

// TODO: Check this implementation
func (r *PostgresUserRepository) InsertUser(user model.User) error {
	query := `INSERT INTO users(name, city, state, country) VALUES($1, $2, $3, $4)`
	if _, err := r.db.Exec(query, user.Name, user.City, user.State, user.Country); err != nil {
		return err
	}

	return nil
}

func (r *PostgresUserRepository) UpdateUser(id int, user model.User) error {
	return nil
}

func (r *PostgresUserRepository) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

func (r *PostgresUserRepository) Count() int {
	query := `SELECT COUNT(*) FROM users;`
	var count int
	if err := r.db.QueryRow(query).Scan(&count); err != nil {
		return 0
	}

	return count
}
