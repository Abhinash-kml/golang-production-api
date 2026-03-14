package repository

import (
	"database/sql"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	oteltracer "go.opentelemetry.io/otel/trace"
)

type PostgresCommentRepository struct {
	db     *sql.DB
	tracer oteltracer.Tracer
}

func NewPostgresCommentRepository(connection *connections.PostgresConnection, tracer oteltracer.Tracer) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: connection.DB, tracer: tracer}
}

func (r *PostgresCommentRepository) Setup() error {
	return nil
}

func (r *PostgresCommentRepository) GetComments() ([]model.Comment, error) {
	query := `SELECT * FROM comments;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	var comment model.Comment

	for rows.Next() {
		rows.Scan(&comment.Id, &comment.AuthorID, &comment.PostId, &comment.Body, &comment.Likes)
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresCommentRepository) GetById(id int) (*model.Comment, error) {
	query := `SELECT * FROM comments WHERE id = $1;`
	var comment model.Comment
	if err := r.db.QueryRow(query, id).Scan(&comment.Id, &comment.AuthorID, &comment.PostId, &comment.Body, &comment.Likes); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *PostgresCommentRepository) GetCommentsOfPost(id int) ([]model.Comment, error) {
	query := `SELECT * FROM comments WHERE postid = $1;`
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	var comment model.Comment

	for rows.Next() {
		rows.Scan(&comment.Id, &comment.AuthorID, &comment.PostId, &comment.Body, &comment.Likes)
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresCommentRepository) InsertComment(comment model.Comment) error {
	query := `INSERT INTO comments(commenterid, postid, body, likes) VALUES($1, $2, $3, $4);`
	if _, err := r.db.Exec(query, comment.AuthorID, comment.PostId, comment.Body, comment.Likes); err != nil {
		return err
	}

	return nil
}

func (r *PostgresCommentRepository) DeleteComment(id int) error {
	query := `DELETE FROM comments WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

// TODO: Implement as per Json Merge patch
func (r *PostgresCommentRepository) UpdateComment(id int, comment model.Comment) error {
	return nil
}

func (r *PostgresCommentRepository) Count() int {
	query := `SELECT COUNT(*) FROM comments;`
	var count int
	if err := r.db.QueryRow(query).Scan(&count); err != nil {
		return 0
	}

	return count
}
