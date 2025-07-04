package repository

import (
	"database/sql"
	"nik-mLb/forum_task.com/internal/models"
)

type ServiceRepo struct {
	db *sql.DB
}

func NewServiceRepo(db *sql.DB) *ServiceRepo{
	return &ServiceRepo{db: db,}
}

type IServiceRepo interface {
	Clear() error
	SelectStatus() (*models.Status ,error)
}

const (
	ClearQuery = `TRUNCATE forum.vote, forum.post, forum.thread, forum.forum, forum.user RESTART IDENTITY CASCADE;`
	StatusQuery =  `SELECT 
    (SELECT COALESCE(SUM(posts), 0) FROM forum WHERE posts > 0) AS post, 
    (SELECT COALESCE(SUM(threads), 0) FROM forum WHERE threads > 0) AS thread, 
    (SELECT COUNT(*) FROM "user") AS user, 
    (SELECT COUNT(*) FROM forum) AS forum;`
)

func (repo *ServiceRepo) Clear() error {
	_, err := repo.db.Query(ClearQuery)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ServiceRepo) SelectStatus() (*models.Status ,error) {
	status := &models.Status{}

	err := repo.db.QueryRow(StatusQuery).Scan(&status.Post, &status.Thread, &status.User, &status.Forum)
	if err != nil {
		return nil, err
	}

	return status, nil
}