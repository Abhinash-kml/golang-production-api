package repository

import (
	"context"
	"database/sql"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	oteltracer "go.opentelemetry.io/otel/trace"
)

type PostgresPostRepository struct {
	db     *sql.DB
	tracer oteltracer.Tracer
}

func NewPostgresPostRepository(connection *connections.PostgresConnection, tracer oteltracer.Tracer) *PostgresPostRepository {
	return &PostgresPostRepository{db: connection.DB, tracer: tracer}
}

func (r *PostgresPostRepository) Setup() error {
	return nil
}

func (r *PostgresPostRepository) GetPosts(ctx context.Context) ([]model.Post, error) {
	ctx, span := r.tracer.Start(ctx, "GetPosts.Repository")
	defer span.End()

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

func (r *PostgresPostRepository) GetById(ctx context.Context, id int) (*model.Post, error) {
	ctx, span := r.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	query := `SELECT * FROM posts WHERE id = $1;`
	var post model.Post
	if err := r.db.QueryRow(query, id).Scan(&post.Id, &post.Title, &post.Body, &post.AuthorID, &post.CreatedAt, &post.Likes); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostgresPostRepository) GetPostsOfUser(ctx context.Context, id int) ([]model.Post, error) {
	ctx, span := r.tracer.Start(ctx, "GetPostsOfUser.Repository")
	defer span.End()

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

func (r *PostgresPostRepository) InsertPost(ctx context.Context, post model.Post) error {
	ctx, span := r.tracer.Start(ctx, "InsertPost.Repository")
	defer span.End()

	query := `INSERT INTO posts(title, body, creatorid, createdat, likes) VALUES($1, $2, $3, $4, $5);`
	if _, err := r.db.Exec(query, post.Title, post.Body, post.AuthorID, post.CreatedAt, post.Likes); err != nil {
		return err
	}

	return nil
}

func (r *PostgresPostRepository) DeletePost(ctx context.Context, id int) error {
	ctx, span := r.tracer.Start(ctx, "DeletePost.Repository")
	defer span.End()

	query := `DELETE FROM posts WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

// TODO: Implement as per JSON Merge Patch
func (r *PostgresPostRepository) UpdatePost(ctx context.Context, id int, post model.Post) error {
	ctx, span := r.tracer.Start(ctx, "UpdatePost.Repository")
	defer span.End()

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
