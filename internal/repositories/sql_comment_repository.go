package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	"go.opentelemetry.io/otel/attribute"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type PostgresCommentRepository struct {
	db     *sql.DB
	tracer oteltracer.Tracer
}

func NewPostgresCommentRepository(connection *connections.PostgresConnection, tracer oteltracer.Tracer) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: connection.DB, tracer: tracer}
}

func (r *PostgresCommentRepository) Setup() error {
	query := `INSERT INTO comments(id, author_id, post_id, body, likes) VALUES($1, $2, $3, $4, $5)
				ON CONFLICT(id)
				DO UPDATE SET
					body = EXCLUDED.body,
					likes = EXCLUDED.likes;`

	file, err := os.OpenFile("./mocks/comments.json", os.O_RDONLY, 0644)
	if err != nil {
		zap.L().Fatal("Failed to open mocks file for comment repository setup")
	}

	var comments []model.Comment
	comments = make([]model.Comment, 150)

	err = json.NewDecoder(file).Decode(&comments)
	if err != nil {
		zap.L().Fatal("Failed to decode json from mocks file", zap.Error(err), zap.String("file", "comments.json"))
	}

	for index := range comments {
		_, err := r.db.Exec(query, comments[index].Id, comments[index].AuthorID, comments[index].PostId, comments[index].Body, comments[index].Likes)
		if err != nil {
			zap.L().Fatal("Failed to execute sql query", zap.Error(err))
		}
	}

	return nil
}

func (r *PostgresCommentRepository) GetComments(ctx context.Context) ([]model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "GetComments.Repository")
	defer span.End()

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

func (r *PostgresCommentRepository) GetById(ctx context.Context, id int) (*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	query := `SELECT * FROM comments WHERE id = $1;`
	var comment model.Comment
	if err := r.db.QueryRow(query, id).Scan(&comment.Id, &comment.AuthorID, &comment.PostId, &comment.Body, &comment.Likes); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *PostgresCommentRepository) GetCommentsOfPost(ctx context.Context, id int) ([]model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "GetCommentsOfPost.Repository")
	defer span.End()

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

func (r *PostgresCommentRepository) InsertComment(ctx context.Context, comment model.Comment) error {
	ctx, span := r.tracer.Start(ctx, "InsertComment.Repository")
	defer span.End()

	query := `INSERT INTO comments(commenterid, postid, body, likes) VALUES($1, $2, $3, $4);`
	if _, err := r.db.Exec(query, comment.AuthorID, comment.PostId, comment.Body, comment.Likes); err != nil {
		return err
	}

	return nil
}

func (r *PostgresCommentRepository) DeleteComment(ctx context.Context, id int) error {
	ctx, span := r.tracer.Start(ctx, "DeleteComment.Repository")
	defer span.End()

	query := `DELETE FROM comments WHERE id = $1;`
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

// TODO: Implement as per Json Merge patch
func (r *PostgresCommentRepository) UpdateComment(ctx context.Context, dto model.CommentUpdateDTO) error {
	ctx, span := r.tracer.Start(ctx, "UpdateComment.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", dto.Id),
		attribute.Int("post.patch.num", len(dto.Patches)))

	// Read fields from patches
	fields := make(map[string]any)
	for index := range dto.Patches {
		current := dto.Patches[index]
		fields[current.Field] = current.Value
	}

	// Build one update query to prevent multiple calls to db
	squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := squirrel.Update("comments").SetMap(fields).Where(squirrel.Eq{"id": dto.Id})
	sqlString, args, _ := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sqlString)

	// Execute the update call
	_, err := r.db.Exec(sqlString, args...)
	if err != nil {
		zap.L().Fatal("Failed to execute sql query", zap.Error(err))
	}

	return nil
}

func (e *PostgresCommentRepository) ReplaceComment(ctx context.Context, dto model.CommentReplaceDTO) error {
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
