package repository

import (
	"errors"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
)

var (
	ErrNoComments         = errors.New("No user in repository")
	ErrUndefinedComments  = errors.New("Undefined users")
	ErrCommentDoesntExist = errors.New("Provided users doesn't exist")
)

type CommentRepository interface {
	Setup() error

	GetComments() ([]model.Comment, error)
	InsertComments([]model.Comment) error
	DeleteComments([]model.Comment) error
	UpdateComments([]model.Comment) error
}

type InMemoryCommentRepository struct {
	comments []model.Comment
}

func NewInMemoryCommentsRepository() *InMemoryCommentRepository {
	return &InMemoryCommentRepository{}
}

func (e *InMemoryCommentRepository) Setup() error {
	comments := []model.Comment{
		// Post 1
		{
			Id:          1,
			CommenterId: 1,
			PostId:      1,
			Body:        "This article explains the problem very clearly.\nI especially liked the examples.",
			Likes:       6,
		},
		{
			Id:          2,
			CommenterId: 2,
			PostId:      1,
			Body:        "I ran into the same issue last week.\nGood to know there’s a cleaner solution.",
			Likes:       4,
		},

		// Post 2
		{
			Id:          3,
			CommenterId: 3,
			PostId:      2,
			Body:        "The approach makes sense, but I think edge cases need more coverage.",
			Likes:       2,
		},
		{
			Id:          4,
			CommenterId: 4,
			PostId:      2,
			Body:        "Can you share benchmarks for this?\nCurious how it behaves under load.",
			Likes:       5,
		},
		{
			Id:          5,
			CommenterId: 5,
			PostId:      2,
			Body:        "Tried this in my own codebase and it worked well.\nThanks for sharing.",
			Likes:       7,
		},

		// Post 3
		{
			Id:          6,
			CommenterId: 6,
			PostId:      3,
			Body:        "Simple explanation and straight to the point.",
			Likes:       3,
		},

		// Post 4
		{
			Id:          7,
			CommenterId: 7,
			PostId:      4,
			Body:        "This cleared up a lot of confusion for me.\nReally appreciate the breakdown.",
			Likes:       8,
		},
		{
			Id:          8,
			CommenterId: 8,
			PostId:      4,
			Body:        "One thing to watch out for is memory usage.\nOtherwise, solid advice.",
			Likes:       4,
		},
		{
			Id:          9,
			CommenterId: 9,
			PostId:      4,
			Body:        "Do you recommend this approach for production systems?",
			Likes:       2,
		},
		{
			Id:          10,
			CommenterId: 10,
			PostId:      4,
			Body:        "Nice write-up.\nVery easy to follow.",
			Likes:       6,
		},

		// Post 5
		{
			Id:          11,
			CommenterId: 11,
			PostId:      5,
			Body:        "I like how you compared this with alternative solutions.",
			Likes:       5,
		},
		{
			Id:          12,
			CommenterId: 12,
			PostId:      5,
			Body:        "This saved me a lot of trial and error.\nThanks!",
			Likes:       9,
		},

		// Post 6
		{
			Id:          13,
			CommenterId: 13,
			PostId:      6,
			Body:        "The diagrams really helped understanding the flow.",
			Likes:       4,
		},
		{
			Id:          14,
			CommenterId: 14,
			PostId:      6,
			Body:        "I’m not fully convinced about the scalability aspect.\nWould love more details.",
			Likes:       3,
		},
		{
			Id:          15,
			CommenterId: 15,
			PostId:      6,
			Body:        "Tested this on a small service and it behaves as expected.",
			Likes:       6,
		},

		// Post 7
		{
			Id:          16,
			CommenterId: 16,
			PostId:      7,
			Body:        "Very practical advice.\nThis matches what we do internally.",
			Likes:       7,
		},

		// Post 8
		{
			Id:          17,
			CommenterId: 17,
			PostId:      8,
			Body:        "Good balance between theory and implementation details.",
			Likes:       5,
		},
		{
			Id:          18,
			CommenterId: 18,
			PostId:      8,
			Body:        "Would be nice to see a real-world example next time.",
			Likes:       3,
		},
		{
			Id:          19,
			CommenterId: 19,
			PostId:      8,
			Body:        "Clean and readable.\nBookmarked for future reference.",
			Likes:       8,
		},

		// Post 9
		{
			Id:          20,
			CommenterId: 20,
			PostId:      9,
			Body:        "This aligns with the patterns I’ve seen in large systems.",
			Likes:       6,
		},
		{
			Id:          21,
			CommenterId: 21,
			PostId:      9,
			Body:        "Minor typo aside, the content is solid.",
			Likes:       2,
		},

		// Post 10
		{
			Id:          22,
			CommenterId: 22,
			PostId:      10,
			Body:        "Excellent conclusion.\nSummarizes everything nicely.",
			Likes:       9,
		},
		{
			Id:          23,
			CommenterId: 23,
			PostId:      10,
			Body:        "I followed along step by step and it worked perfectly.",
			Likes:       7,
		},
		{
			Id:          24,
			CommenterId: 24,
			PostId:      10,
			Body:        "Looking forward to the next post in this series.",
			Likes:       5,
		},
	}

	// Assign fake to real container ( this is safe because of heap escape )
	e.comments = comments

	return nil
}

func (e *InMemoryCommentRepository) GetComments() ([]model.Comment, error) {
	if len(e.comments) <= 0 {
		return nil, ErrNoComments
	}

	return e.comments, nil
}

func (e *InMemoryCommentRepository) InsertComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	e.comments = append(e.comments, comments...)
	return nil
}

func (e *InMemoryCommentRepository) DeleteComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	if len(e.comments) <= 0 {
		return ErrNoComments
	}

	e.comments = slices.DeleteFunc(e.comments, func(e model.Comment) bool {
		for _, value := range comments {
			if value.Id == e.Id {
				return true
			}
		}

		return false
	})

	return nil
}

func (e *InMemoryCommentRepository) UpdateComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, internal := range e.comments {
		for _, incoming := range comments {
			if internal.Body == incoming.Body {
				internal.Body = incoming.Body
			}
		}
	}

	return nil
}
