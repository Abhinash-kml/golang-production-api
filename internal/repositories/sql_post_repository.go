package repository

import (
	"database/sql"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
)

type PostgresPostRepository struct {
	db *sql.DB
}

func NewPostgresPostRepository(connection *connections.PostgresConnection) *PostgresPostRepository {
	return &PostgresPostRepository{db: connection.DB}
}

func (r *PostgresPostRepository) Setup() error {
	return nil
}

func (r *PostgresPostRepository) GetPosts() ([]model.Post, error) {
	query := `SELECT * FROM posts;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	var post model.Post

	for rows.Next() {
		rows.Scan(&post.Id, &post.Title, &post.Body, &post.AuthorID, &post.CreatedAt, &post.Likes)
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostgresPostRepository) GetById(id int) (*model.Post, error) {
	query := `SELECT * FROM posts WHERE id = $1;`
	var post model.Post
	if err := r.db.QueryRow(query, id).Scan(&post.Id, &post.Title, &post.Body, &post.AuthorID, &post.CreatedAt, &post.Likes); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostgresPostRepository) GetPostsOfUser(id int) ([]model.Post, error) {
	query := `SELECT * FROM posts WHERE creatorid = $1;`
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	var post model.Post

	for rows.Next() {
		rows.Scan(&post.Id, &post.Title, &post.Body, &post.AuthorID, &post.CreatedAt, &post.Likes)
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostgresPostRepository) InsertPost(post model.Post) error {
	query := `INSERT INTO posts(title, body, creatorid, createdat, likes) VALUES($1, $2, $3, $4, $5);`
	if _, err := r.db.Exec(query, post.Title, post.Body, post.AuthorID, post.CreatedAt, post.Likes); err != nil {
		return err
	}

	return nil
}

func (r *PostgresPostRepository) DeletePost(id int) error {
	query := `DELETE FROM posts WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

// TODO: Implement as per JSON Merge Patch
func (r *PostgresPostRepository) UpdatePost(id int, post model.Post) error {
	return nil
}

func (r *PostgresPostRepository) Count() int {
	query := `SELECT COUNT(*) FROM posts;`
	var count int
	if err := r.db.QueryRow(query).Scan(&count); err != nil {
		return 0
	}

	return count
}
