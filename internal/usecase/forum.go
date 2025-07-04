package usecase

import (
	"errors"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/repository"
	"sync"
	"time"
)

type ForumUseCase struct {
    forumRepo  repository.IForumRepository
    userRepo   repository.IUserRepository
}

func NewForumUseCase(forumRepo repository.IForumRepository, userRepo repository.IUserRepository) *ForumUseCase {
    return &ForumUseCase{
        forumRepo:  forumRepo,
        userRepo:   userRepo,
    }
}

type IForumUseCase interface {
    CreateForum(forum models.Forum) (*models.Forum, error)
    GetForumBySlug(slug string) (*models.Forum, error)
    CreateThread(slug string, thread models.Thread) (*models.Thread, error)
    GetForumUsers(slug string, limit int, since string, desc bool) ([]models.User, error)
    GetForumThreads(slug string, limit int, since string, desc bool) ([]models.Thread, error)
}

func (uc *ForumUseCase) CreateForum(forum models.Forum) (*models.Forum, error) {
    var wg sync.WaitGroup
    var userErr, forumErr error
    var user *models.User
    var existingForum *models.Forum

    wg.Add(2)

    // Проверяем существование пользователя
    go func() {
        defer wg.Done()
        user, userErr = uc.userRepo.GetUserByNickname(forum.User)
    }()

    // Проверяем существование форума
    go func() {
        defer wg.Done()
        existingForum, forumErr = uc.forumRepo.GetForumBySlug(forum.Slug)
    }()

    wg.Wait()

    if userErr != nil {
        return nil, errors.New("user not found")
    }
    forum.User = user.Nickname

    if forumErr == nil && existingForum != nil {
        return existingForum, errors.New("forum already exists")
    }

    // Создаем форум
    forum.Posts = 0
    forum.Threads = 0
    err := uc.forumRepo.CreateForum(forum)
    if err != nil {
        return nil, err
    }

    return &forum, nil
}

func (uc *ForumUseCase) GetForumBySlug(slug string) (*models.Forum, error) {
    return uc.forumRepo.GetForumBySlug(slug)
}

func (uc *ForumUseCase) CreateThread(slug string, thread models.Thread) (*models.Thread, error) {
    var wg sync.WaitGroup
    var forumErr, userErr, threadErr error
    var forum *models.Forum
    var existingThread *models.Thread

    wg.Add(2)

    go func() {
        defer wg.Done()
        forum, forumErr = uc.forumRepo.GetForumBySlug(slug)
    }()

    // Проверяем существование пользователя
    go func() {
        defer wg.Done()
        _, userErr = uc.userRepo.GetUserByNickname(thread.Author)
    }()


    // Проверяем существование ветки с таким slug
    if thread.Slug != "" {
        wg.Add(1)
        go func() {
            defer wg.Done()
            existingThread, threadErr = uc.forumRepo.GetThreadBySlug(thread.Slug)
        }()
    }

    wg.Wait() // Ждём завершения всех горутин

    if forumErr != nil {
        return nil, errors.New("forum not found")
    }
    if userErr != nil {
        return nil, errors.New("user not found")
    }
    if threadErr == nil && existingThread != nil {
        return existingThread, errors.New("thread already exists")
    }

    // Устанавливаем оставшиеся поля
    thread.Forum = forum.Slug
    if thread.Created.IsZero() {
        thread.Created = time.Now()
    }

    err := uc.forumRepo.CreateThread(&thread)
    if err != nil {
        return nil, err
    }

    return &thread, nil
}

func (uc *ForumUseCase) GetForumThreads(slug string, limit int, since string, desc bool) ([]models.Thread, error) {
    // Проверяем существование форума
    _, err := uc.forumRepo.GetForumBySlug(slug)
    if err != nil {
        return nil, errors.New("forum not found")
    }
    
    return uc.forumRepo.GetForumThreads(slug, limit, since, desc)
}

func (uc *ForumUseCase) GetForumUsers(slug string, limit int, since string, desc bool) ([]models.User, error) {
    // Проверяем существование форума
    if _, err := uc.forumRepo.GetForumBySlug(slug); err != nil {
        return nil, errors.New("forum not found")
    }

    users, err := uc.forumRepo.GetForumUsers(slug, limit, since, desc)
    if err != nil {
        return nil, err
    }

    return users, nil
}