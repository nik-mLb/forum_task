package repository

import (
    "database/sql"
    "errors"
    "nik-mLb/forum_task.com/internal/models"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

type IUserRepository interface {
    CreateUser(user models.User) error
    GetUserByNickname(nickname string) (*models.User, error)
    GetUsersByNicknameOrEmail(nickname, email string) ([]*models.User, error)
    UpdateUser(user models.User) error
}

const (
    createUserQuery = `
        INSERT INTO forum."user" (nickname, fullname, about, email) 
        VALUES ($1, $2, $3, $4)`
    
    getUserByNicknameQuery = `
        SELECT nickname, fullname, about, email 
        FROM forum."user" 
        WHERE LOWER(nickname) = LOWER($1)`
    
    getUsersByNicknameOrEmailQuery = `
        SELECT nickname, fullname, about, email 
        FROM forum."user" 
        WHERE LOWER(nickname) = LOWER($1) OR LOWER(email) = LOWER($2)`
    
    updateUserQuery = `
        UPDATE forum."user" 
        SET fullname = COALESCE(NULLIF($2, ''), fullname),
            about = COALESCE(NULLIF($3, ''), about),
            email = COALESCE(NULLIF($4, ''), email)
        WHERE nickname = $1
        RETURNING nickname, fullname, about, email`
)

func (r *UserRepository) CreateUser(user models.User) error {
    _, err := r.db.Exec(
        createUserQuery,
        user.Nickname,
        user.Fullname,
        user.About,
        user.Email,
    )
    return err
}

func (r *UserRepository) GetUserByNickname(nickname string) (*models.User, error) {
    user := &models.User{}
    err := r.db.QueryRow(
        getUserByNicknameQuery,
        nickname,
    ).Scan(
        &user.Nickname,
        &user.Fullname,
        &user.About,
        &user.Email,
    )
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return user, nil
}

func (r *UserRepository) GetUsersByNicknameOrEmail(nickname, email string) ([]*models.User, error) {
    rows, err := r.db.Query(
        getUsersByNicknameOrEmailQuery,
        nickname,
        email,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*models.User
    for rows.Next() {
        user := &models.User{}
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

    return users, nil
}

func (r *UserRepository) UpdateUser(user models.User) error {
    updatedUser := &models.User{}
    err := r.db.QueryRow(
        updateUserQuery,
        user.Nickname,
        user.Fullname,
        user.About,
        user.Email,
    ).Scan(
        &updatedUser.Nickname,
        &updatedUser.Fullname,
        &updatedUser.About,
        &updatedUser.Email,
    )
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return errors.New("user not found")
        }
        return err
    }
    return nil
}