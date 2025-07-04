package repository

import (
	"database/sql"
	"errors"
	"nik-mLb/forum_task.com/internal/models"
)

type PostRepository struct {
    db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
    return &PostRepository{db: db}
}

type IPostRepository interface {
	GetPostByID(id int) (*models.Post, error)
	UpdatePost(post *models.Post) (*models.Post, error)
}

const (
	getPostByIDQuery = `
		SELECT id, parent, author, message, isEdited, forum, thread, created, path
		FROM forum.post
		WHERE id = $1`

	updatePostQuery = `
		UPDATE forum.post
		SET message = $1, isEdited = $2
		WHERE id = $3
		RETURNING id, parent, author, message, isEdited, forum, thread, created, path`
)

func (r *PostRepository) GetPostByID(id int) (*models.Post, error) {
	post := &models.Post{}
	err := r.db.QueryRow(
		getPostByIDQuery,
		id,
	).Scan(
		&post.ID,
		&post.Parent,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&post.Thread,
		&post.Created,
		&post.Path,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return post, nil
}

func (r *PostRepository) UpdatePost(post *models.Post) (*models.Post, error) {
	updatedPost := &models.Post{}
	err := r.db.QueryRow(
		updatePostQuery,
		post.Message,
		post.IsEdited,
		post.ID,
	).Scan(
		&updatedPost.ID,
		&updatedPost.Parent,
		&updatedPost.Author,
		&updatedPost.Message,
		&updatedPost.IsEdited,
		&updatedPost.Forum,
		&updatedPost.Thread,
		&updatedPost.Created,
		&updatedPost.Path,
	)

	if err != nil {
		return nil, err
	}
	return updatedPost, nil
}