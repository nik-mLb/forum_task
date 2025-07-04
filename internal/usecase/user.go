package usecase

import (
    "errors"
    "nik-mLb/forum_task.com/internal/models"
    "nik-mLb/forum_task.com/internal/repository"
)

type UserUseCase struct {
    repo repository.IUserRepository
}

func NewUserUseCase(repo repository.IUserRepository) *UserUseCase {
    return &UserUseCase{repo: repo}
}

type IUserUseCase interface {
    CreateUser(newUser models.NewUser, nickname string) (*models.User, error)
    GetUserByNickname(nickname string) (*models.User, error)
    UpdateUser(nickname string, update models.UserUpdate) (*models.User, error)
	GetUserByNicknameOrEmail(nickname string, email string) ([]*models.User, error)
}

func (uc *UserUseCase) CreateUser(newUser models.NewUser, nickname string) (*models.User, error) {
    user := models.User{
        Nickname: nickname,
        Fullname: newUser.Fullname,
        About:    newUser.About,
        Email:    newUser.Email,
    }

    // Проверяем существование пользователя с таким nickname или email
    existingUsers, err := uc.repo.GetUsersByNicknameOrEmail(nickname, newUser.Email)
    if err == nil && len(existingUsers) > 0 {
        return nil, errors.New("user already exists")
    }

    err = uc.repo.CreateUser(user)
    if err != nil {
        return nil, err
    }

    return &user, nil
}

func (uc *UserUseCase) GetUserByNickname(nickname string) (*models.User, error) {
    return uc.repo.GetUserByNickname(nickname)
}

func (uc *UserUseCase) GetUserByNicknameOrEmail(nickname string, email string) ([]*models.User, error) {
    return uc.repo.GetUsersByNicknameOrEmail(nickname, email)
}

func (uc *UserUseCase) UpdateUser(nickname string, update models.UserUpdate) (*models.User, error) {
    // Получаем текущие данные пользователя
    user, err := uc.repo.GetUserByNickname(nickname)
    if err != nil {
        return nil, err
    }

    // Обновляем только те поля, которые пришли в запросе
    if update.Fullname != "" {
        user.Fullname = update.Fullname
    }
    if update.About != "" {
        user.About = update.About
    }
    if update.Email != "" {
        // Проверяем, что новый email не занят другим пользователем
        if update.Email != user.Email {
            existingUsers, err := uc.repo.GetUsersByNicknameOrEmail("", update.Email)
            if err == nil && len(existingUsers) > 0 {
                return nil, errors.New("email already exists")
            }
        }
        user.Email = update.Email
    }

    err = uc.repo.UpdateUser(*user)
    if err != nil {
        return nil, err
    }

    return user, nil
}