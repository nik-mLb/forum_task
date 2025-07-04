package repository

import (
	"database/sql"
	"errors"
	"nik-mLb/forum_task.com/internal/models"
	"strconv"
	"time"
)

type ThreadRepository struct {
    db *sql.DB
}

func NewThreadRepository(db *sql.DB) *ThreadRepository {
    return &ThreadRepository{db: db}
}

type IThreadRepository interface {
	GetThreadBySlugOrID(slugOrID string) (*models.Thread, error)
    CheckPostInThread(postID, threadID int) (bool, error)
    CreatePosts(newPosts models.NewPosts, thread *models.Thread) (models.Posts, error)
	GetThreadPosts(threadID, limit, since int, sort string, desc bool) (models.Posts, error)
	UpdateThread(thread *models.Thread) (*models.Thread, error)
	AddOrUpdateVote(threadID int, vote models.Vote) error
    GetThreadByID(id int) (*models.Thread, error)
}

const (
	getThreadBySlugOrIDQuery = `
        SELECT id, title, author, forum, message, votes, slug, created 
        FROM forum.thread 
        WHERE LOWER(slug) = LOWER($1) OR id = CAST($1 AS INTEGER)`
    
    checkPostInThreadQuery = `
        SELECT EXISTS(
            SELECT 1 FROM forum.post 
            WHERE id = $1 AND thread = $2
        )`
    
    createPostQuery = `
        INSERT INTO forum.post 
        (parent, author, message, isEdited, forum, thread, created, path) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, 
            CASE WHEN $1 = 0 THEN ARRAY[nextval('forum.post_id_seq'::regclass)] 
            ELSE (SELECT path || nextval('forum.post_id_seq'::regclass) FROM forum.post WHERE id = $1) 
            END)
        RETURNING id, parent, author, message, isEdited, forum, thread, created, path`

	getThreadPostsFlatQuery = `
        SELECT id, parent, author, message, isEdited, forum, thread, created 
        FROM forum.post 
        WHERE thread = $1 AND id > $3
        ORDER BY created %s, id %s
        LIMIT $2`
    
    getThreadPostsTreeQuery = `
        SELECT id, parent, author, message, isEdited, forum, thread, created 
        FROM forum.post 
        WHERE thread = $1 AND id > $3
        ORDER BY path %s
        LIMIT $2`
    
    getThreadPostsParentTreeQuery = `
        WITH roots AS (
            SELECT id FROM forum.post 
            WHERE thread = $1 AND parent = 0 AND id > $3
            ORDER BY id %s
            LIMIT $2
        )
        SELECT p.id, p.parent, p.author, p.message, p.isEdited, p.forum, p.thread, p.created 
        FROM forum.post p
        JOIN roots ON p.path[1] = roots.id
        ORDER BY roots.id %s, p.path %s`

	updateThreadQuery = `
        UPDATE forum.thread 
        SET title = $1, message = $2 
        WHERE id = $3
        RETURNING id, title, author, forum, message, votes, slug, created`

	addOrUpdateVoteQuery = `
        INSERT INTO forum.vote (nickname, thread, voice) 
        VALUES ($1, $2, $3)
        ON CONFLICT (nickname, thread) DO UPDATE 
        SET voice = EXCLUDED.voice`
    
    getThreadByIDQuery = `
        SELECT id, title, author, forum, message, votes, slug, created 
        FROM forum.thread 
        WHERE id = $1`
)

func (r *ThreadRepository) GetThreadBySlugOrID(slugOrID string) (*models.Thread, error) {
    thread := &models.Thread{}
    
    // Сначала пробуем найти по ID
    if id, err := strconv.Atoi(slugOrID); err == nil {
        err := r.db.QueryRow(
            "SELECT id, title, author, forum, message, votes, slug, created FROM forum.thread WHERE id = $1",
            id,
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
        
        if err == nil {
            return thread, nil
        }
    }
    
    // Если не нашли по ID или параметр не число, ищем по slug
    err := r.db.QueryRow(
        "SELECT id, title, author, forum, message, votes, slug, created FROM forum.thread WHERE LOWER(slug) = LOWER($1)",
        slugOrID,
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
        return nil, err
    }
    return thread, nil
}

func (r *ThreadRepository) CheckPostInThread(postID, threadID int) (bool, error) {
    var exists bool
    err := r.db.QueryRow(checkPostInThreadQuery, postID, threadID).Scan(&exists)
    return exists, err
}

func (r *ThreadRepository) CreatePosts(newPosts models.NewPosts, thread *models.Thread) (models.Posts, error) {
    tx, err := r.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO forum.post 
        (parent, author, message, isEdited, forum, thread, created) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, parent, author, message, isEdited, forum, thread, created, path`)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    createdPosts := make(models.Posts, 0, len(newPosts))
    createdAt := time.Now()

    for _, newPost := range newPosts {
        var post models.Post
        err := stmt.QueryRow(
            newPost.Parent,
            newPost.Author,
            newPost.Message,
            false,
            thread.Forum,
            thread.ID,
            createdAt,
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
            return nil, err
        }
        createdPosts = append(createdPosts, &post)
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return createdPosts, nil
}

func (r *ThreadRepository) GetThreadPosts(threadID, limit, since int, sort string, desc bool) (models.Posts, error) {
    var query string
    var rows *sql.Rows
    var err error

    order := "ASC"
    if desc {
        order = "DESC"
    }

    switch sort {
    case "flat":
        query = `
            SELECT id, parent, author, message, isEdited, forum, thread, created 
            FROM forum.post 
            WHERE thread = $1`
        
        if since > 0 {
            if desc {
                query += " AND id < $2"
            } else {
                query += " AND id > $2"
            }
            query += " ORDER BY created " + order + ", id " + order
            query += " LIMIT $3"
            rows, err = r.db.Query(query, threadID, since, limit)
        } else {
            query += " ORDER BY created " + order + ", id " + order
            query += " LIMIT $2"
            rows, err = r.db.Query(query, threadID, limit)
        }

    case "tree":
        query = `
            SELECT id, parent, author, message, isEdited, forum, thread, created 
            FROM forum.post 
            WHERE thread = $1`
        
        if since > 0 {
            if desc {
                // Для tree сортировки при desc=true ищем path < текущего
                query += " AND path < (SELECT path FROM forum.post WHERE id = $2)"
            } else {
                // Для tree сортировки при desc=false ищем path > текущего
                query += " AND path > (SELECT path FROM forum.post WHERE id = $2)"
            }
            query += " ORDER BY path " + order
            query += " LIMIT $3"
            rows, err = r.db.Query(query, threadID, since, limit)
        } else {
            query += " ORDER BY path " + order
            query += " LIMIT $2"
            rows, err = r.db.Query(query, threadID, limit)
        }

    case "parent_tree":
        if since > 0 {
            compOp := ">"
            orderRoots := "ASC"
            orderRootsInResult := "ASC"
            if desc {
                compOp = "<"
                orderRoots = "DESC"
                orderRootsInResult = "DESC"
            }

            query = `
                WITH roots AS (
                    SELECT id 
                    FROM forum.post 
                    WHERE thread = $1 AND parent = 0 
                    AND id ` + compOp + ` (SELECT path[1] FROM forum.post WHERE id = $3)
                    ORDER BY id ` + orderRoots + `
                    LIMIT $2
                )
                SELECT p.id, p.parent, p.author, p.message, p.isEdited, p.forum, p.thread, p.created 
                FROM forum.post p
                JOIN roots ON p.path[1] = roots.id
                ORDER BY roots.id ` + orderRootsInResult + `, p.path`
            rows, err = r.db.Query(query, threadID, limit, since)
        } else {
            orderRoots := "ASC"
            orderRootsInResult := "ASC"
            if desc {
                orderRoots = "DESC"
                orderRootsInResult = "DESC"
            }
            
            query = `
                WITH roots AS (
                    SELECT id 
                    FROM forum.post 
                    WHERE thread = $1 AND parent = 0
                    ORDER BY id ` + orderRoots + `
                    LIMIT $2
                )
                SELECT p.id, p.parent, p.author, p.message, p.isEdited, p.forum, p.thread, p.created 
                FROM forum.post p
                JOIN roots ON p.path[1] = roots.id
                ORDER BY roots.id ` + orderRootsInResult + `, p.path`
            rows, err = r.db.Query(query, threadID, limit)
        }

    default:
        return nil, errors.New("invalid sort type")
    }

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    posts := make(models.Posts, 0)
    for rows.Next() {
        var post models.Post
        err := rows.Scan(
            &post.ID,
            &post.Parent,
            &post.Author,
            &post.Message,
            &post.IsEdited,
            &post.Forum,
            &post.Thread,
            &post.Created,
        )
        if err != nil {
            return nil, err
        }
        posts = append(posts, &post)
    }

    return posts, nil
}

func (r *ThreadRepository) UpdateThread(thread *models.Thread) (*models.Thread, error) {
    updatedThread := &models.Thread{}
    err := r.db.QueryRow(
        updateThreadQuery,
        thread.Title,
        thread.Message,
        thread.ID,
    ).Scan(
        &updatedThread.ID,
        &updatedThread.Title,
        &updatedThread.Author,
        &updatedThread.Forum,
        &updatedThread.Message,
        &updatedThread.Votes,
        &updatedThread.Slug,
        &updatedThread.Created,
    )
    
    if err != nil {
        return nil, err
    }
    return updatedThread, nil
}

func (r *ThreadRepository) AddOrUpdateVote(threadID int, vote models.Vote) error {
    _, err := r.db.Exec(
        addOrUpdateVoteQuery,
        vote.Nickname,
        threadID,
        vote.Voice,
    )
    return err
}

func (r *ThreadRepository) GetThreadByID(id int) (*models.Thread, error) {
    thread := &models.Thread{}
    err := r.db.QueryRow(
        getThreadByIDQuery,
        id,
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
        return nil, err
    }
    return thread, nil
}