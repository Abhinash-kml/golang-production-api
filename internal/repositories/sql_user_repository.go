package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"

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
	query := `INSERT INTO users(id, name, city, state, country) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id)
				DO UPDATE SET
				name = EXCULDED.name,
				city = EXCLUDED.city,
				state = EXCLUDED.state,
				country = EXCULDED.country;`

	file, err := os.OpenFile("./mocks/users.json", os.O_RDONLY, 0644)
	if err != nil {
		zap.L().Fatal("Failed to open file for respitory setup", zap.Error(err), zap.String("file", "users.json"))
	}

	var users []model.User
	users = make([]model.User, 0, 150)
	err = json.NewDecoder(file).Decode(&users)
	if err != nil {
		zap.L().Fatal("Faile dto decode json from mocks file")
	}

	for index := range users {
		_, err := r.db.Exec(query, users[index].Id, users[index].Name, users[index].City, users[index].State, users[index].Country)
		if err != nil {
			zap.L().Fatal("Faile dto execute query", zap.Error(err))
		}
	}

	return nil
}

func (r *PostgresUserRepository) GetUsers(ctx context.Context) ([]model.User, error) {
	ctx, span := r.tracer.Start(ctx, "GetUsers.Repository")
	defer span.End()

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

func (r *PostgresUserRepository) GetById(ctx context.Context, id int) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	query := `SELECT * FROM users WHERE id = $1;`
	var user model.User
	if err := r.db.QueryRow(query, id).Scan(&user.Id, &user.Name, &user.City, &user.State, &user.Country); err != nil {
		return nil, err
	}

	return &user, nil

}

// TODO: Check this implementation
func (r *PostgresUserRepository) InsertUser(ctx context.Context, user model.User) error {
	ctx, span := r.tracer.Start(ctx, "InsertUser.Repository")
	defer span.End()

	query := `INSERT INTO users(name, city, state, country) VALUES($1, $2, $3, $4)`
	if _, err := r.db.Exec(query, user.Name, user.City, user.State, user.Country); err != nil {
		return err
	}

	return nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user model.UserUpdateDTO) error {
	ctx, span := r.tracer.Start(ctx, "UpdateUser.Repository")
	defer span.End()

	return nil
}

func (e *PostgresUserRepository) ReplaceUser(ctx context.Context, dto model.UserReplaceDTO) error {
	return nil
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id int) error {
	ctx, span := r.tracer.Start(ctx, "DeleteUser.Repository")
	defer span.End()

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
