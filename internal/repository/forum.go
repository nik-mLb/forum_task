package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"nik-mLb/forum_task.com/internal/models"
	"strings"
	"time"
)

type ForumRepository struct {
    db *sql.DB
}

func NewForumRepository(db *sql.DB) *ForumRepository {
    return &ForumRepository{db: db}
}

type IForumRepository interface {
    CreateForum(forum models.Forum) error
    GetForumBySlug(slug string) (*models.Forum, error)
    CreateThread(thread *models.Thread) error
    GetThreadBySlug(slug string) (*models.Thread, error)
    GetForumUsers(slug string, limit int, since string, desc bool) ([]models.User, error)
    GetForumThreads(slug string, limit int, since string, desc bool) ([]models.Thread, error)
}

const (
    createForumQuery = `
        INSERT INTO forum.forum (title, "user", slug, posts, threads) 
        VALUES ($1, $2, $3, $4, $5)`
    
    getForumBySlugQuery = `
        SELECT title, "user", slug, posts, threads 
        FROM forum.forum 
        WHERE LOWER(slug) = LOWER($1)`
    
    createThreadQuery = `
        INSERT INTO forum.thread (title, author, forum, message, votes, slug, created) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`
    
    getThreadBySlugQuery = `
        SELECT id, title, author, forum, message, votes, slug, created 
        FROM forum.thread 
        WHERE LOWER(slug) = LOWER($1)`
    
    getForumThreadsQuery = `
        SELECT id, title, author, forum, message, votes, slug, created 
        FROM forum.thread 
        WHERE LOWER(forum) = LOWER($1)`

    getForumUsersQuery = `
        SELECT u.nickname, u.fullname, u.about, u.email 
        FROM forum."user" u
        WHERE u.nickname IN (
            SELECT author FROM forum.post WHERE forum = $1
            UNION
            SELECT author FROM forum.thread WHERE forum = $1
        )`
    
    getForumUsersWithSinceQuery = `
        SELECT u.nickname, u.fullname, u.about, u.email 
        FROM forum."user" u
        WHERE u.nickname IN (
            SELECT author FROM forum.post WHERE forum = $1
            UNION
            SELECT author FROM forum.thread WHERE forum = $1
        ) AND LOWER(u.nickname) %s LOWER($2)`
    
    getForumUsersOrderQuery = ` ORDER BY LOWER(u.nickname) %s LIMIT $%d`
)

func (r *ForumRepository) CreateForum(forum models.Forum) error {
    _, err := r.db.Exec(
        createForumQuery,
        forum.Title,
        forum.User,
        forum.Slug,
        forum.Posts,
        forum.Threads,
    )
    return err
}

func (r *ForumRepository) GetForumBySlug(slug string) (*models.Forum, error) {
    forum := &models.Forum{}
    err := r.db.QueryRow(
        getForumBySlugQuery,
        slug,
    ).Scan(
        &forum.Title,
        &forum.User,
        &forum.Slug,
        &forum.Posts,
        &forum.Threads,
    )
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.New("forum not found")
        }
        return nil, err
    }
    return forum, nil
}

func (r *ForumRepository) CreateThread(thread *models.Thread) error {
    err := r.db.QueryRow(
        createThreadQuery,
        thread.Title,
        thread.Author,
        thread.Forum,
        thread.Message,
        thread.Votes,
        thread.Slug,
        thread.Created,
    ).Scan(&thread.ID)
    
    return err
}

func (r *ForumRepository) GetThreadBySlug(slug string) (*models.Thread, error) {
    thread := &models.Thread{}
    err := r.db.QueryRow(
        getThreadBySlugQuery,
        slug,
    ).Scan(
        &thread.ID,
        &thread.Title,
        &thread.Author,
        &thread.Forum,
        &thread.Message,
        &thread.Votes,
        &thread.Slug,
        &thread.Created,
    )
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.New("thread not found")
        }
        return nil, err
    }
    return thread, nil
}

func (r *ForumRepository) GetForumThreads(slug string, limit int, since string, desc bool) ([]models.Thread, error) {
    threads := make([]models.Thread, 0)
    var rows *sql.Rows
    var err error
    
    order := "ASC"
    if desc {
        order = "DESC"
    }
    
    baseQuery := getForumThreadsQuery
    
    // Добавляем условие для since с учетом сортировки
    if since != "" {
        _, err = time.Parse(time.RFC3339, since)
        if err != nil {
            return nil, err
        }
        
        if desc {
            baseQuery += " AND created <= $3"
        } else {
            baseQuery += " AND created >= $3"
        }
    }

    baseQuery += " ORDER BY created " + order
    
    // Добавляем лимит
    baseQuery += " LIMIT $2"
    
    // Выполняем запрос
    if since != "" {
        sinceTime, _ := time.Parse(time.RFC3339, since)
        rows, err = r.db.Query(baseQuery, slug, limit, sinceTime)
    } else {
        rows, err = r.db.Query(baseQuery, slug, limit)
    }
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var thread models.Thread
        err := rows.Scan(
            &thread.ID,
            &thread.Title,
            &thread.Author,
            &thread.Forum,
            &thread.Message,
            &thread.Votes,
            &thread.Slug,
            &thread.Created,
        )
        if err != nil {
            return nil, err
        }
        threads = append(threads, thread)
    }
    
    if err = rows.Err(); err != nil {
        return nil, err
    }
    
    return threads, nil
}

func (r *ForumRepository) GetForumUsers(slug string, limit int, since string, desc bool) ([]models.User, error) {
    params := []interface{}{strings.ToLower(slug)}

    baseQuery := `
        SELECT u.nickname, u.fullname, u.about, u.email 
        FROM forum."user" u
        WHERE u.nickname IN (
            SELECT author FROM forum.post WHERE LOWER(forum) = $1
            UNION
            SELECT author FROM forum.thread WHERE LOWER(forum) = $1
        )`

    // Добавляем условие для since
    if since != "" {
        comparison := ">"
        if desc {
            comparison = "<"
        }
        baseQuery += fmt.Sprintf(" AND LOWER(u.nickname) %s LOWER($2)", comparison)
        params = append(params, since)
    }

    order := "ASC"
    if desc {
        order = "DESC"
    }
    baseQuery += fmt.Sprintf(" ORDER BY LOWER(u.nickname) %s", order)

    // Добавляем лимит
    paramIndex := len(params) + 1
    baseQuery += fmt.Sprintf(" LIMIT $%d", paramIndex)
    params = append(params, limit)

    rows, err := r.db.Query(baseQuery, params...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    users := make([]models.User, 0)
    for rows.Next() {
        var user models.User
        if err := rows.Scan(
            &user.Nickname,
            &user.Fullname,
            &user.About,
            &user.Email,
        ); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, rows.Err()
}

// Вспомогательная функция для порядка сортировки
func getOrder(desc bool) string {
    if desc {
        return "DESC"
    }
    return "ASC"
}