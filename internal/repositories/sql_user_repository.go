package repository

import (
	"database/sql"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type SqlUsersRepository struct {
	db *sql.DB
}

func (r *SqlUsersRepository) Setup() error {
	db, err := sql.Open("postgres", "postgresql://postgres:Abx305@localhost:5432")
	if err != nil {
		zap.L().Fatal("SQL Connection failed", zap.Error(err))
	}

	err = db.Ping()
	if err != nil {
		zap.L().Warn("SQL Database ping issue", zap.Error(err))
	}

	zap.L().Info("Connection to Postgres successful")
	r.db = db

	return nil
}
func (r *SqlUsersRepository) GetUsers() ([]model.User, error) {
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
func (r *SqlUsersRepository) GetById(id int) (*model.User, error) {
	query := `SELECT * FROM users WHERE id = $1;`
	var user model.User
	if err := r.db.QueryRow(query, id).Scan(&user.Id, &user.Name, &user.City, &user.State, &user.Country); err != nil {
		return nil, err
	}

	return &user, nil

}

// TODO: Check this implementation
func (r *SqlUsersRepository) InsertUser(user model.User) error {
	query := `INSERT INTO users VALUES($1, $2, $3, $4)`
	if _, err := r.db.Exec(query, user.Name, user.City, user.State, user.Country); err != nil {
		return err
	}

	return nil
}
func (r *SqlUsersRepository) UpdateUser(id int, user model.User) error {
	return nil
}

func (r *SqlUsersRepository) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

func (r *SqlUsersRepository) Count() int {
	query := `SELECT COUNT(*) FROM users;`
	var count int
	if err := r.db.QueryRow(query).Scan(&count); err != nil {
		return 0
	}

	return count
}
