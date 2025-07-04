package usecase

import (
	"errors"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/repository"
	"sync"
)

type ThreadUseCase struct {
    threadRepo  repository.IThreadRepository
	userRepo repository.IUserRepository
}

func NewThreadUseCase(threadRepo repository.IThreadRepository, userRepo repository.IUserRepository) *ThreadUseCase {
    return &ThreadUseCase{
        threadRepo:  threadRepo,
		userRepo: userRepo,
    }
}

type IThreadUseCase interface {
	CreatePosts(slugOrID string, newPosts models.NewPosts) (models.Posts, error)
	GetThreadPosts(slugOrID string, limit, since int, sort string, desc bool) (models.Posts, error)
	GetThreadDetails(slugOrID string) (*models.Thread, error)
    UpdateThread(slugOrID string, update models.ThreadUpdate) (*models.Thread, error)
	VoteForThread(slugOrID string, vote models.Vote) (*models.Thread, error)
}

func (uc *ThreadUseCase) CreatePosts(slugOrID string, newPosts models.NewPosts) (models.Posts, error) {
    // Получаем ветку по slug или ID
    thread, err := uc.threadRepo.GetThreadBySlugOrID(slugOrID)
    if err != nil {
        return nil, errors.New("thread not found")
    }

    var (
        wg          sync.WaitGroup
        mu          sync.Mutex
        firstErr    error
        semaphore   = make(chan struct{}, 10) // Ограничиваем кол-во горутин
    )

    // Проверяем авторов и родительские посты
    for _, newPost := range newPosts {
        wg.Add(1)
        go func(post models.NewPost) {
            defer wg.Done()
            semaphore <- struct{}{} // Захватываем слот
            defer func() { <-semaphore }() // Освобождаем

            // Проверка автора
            if _, err := uc.userRepo.GetUserByNickname(post.Author); err != nil {
                mu.Lock()
                if firstErr == nil {
                    firstErr = errors.New("author not found")
                }
                mu.Unlock()
                return
            }

            // Проверка родительского поста
            if post.Parent != 0 {
                exists, err := uc.threadRepo.CheckPostInThread(post.Parent, thread.ID)
                if err != nil || !exists {
                    mu.Lock()
                    if firstErr == nil {
                        firstErr = errors.New("parent post not found")
                    }
                    mu.Unlock()
                    return
                }
            }
        }(*newPost)
    }

    wg.Wait()

    if firstErr != nil {
        return nil, firstErr
    }

    // Создаем посты
    createdPosts, err := uc.threadRepo.CreatePosts(newPosts, thread)
    if err != nil {
        return nil, err
    }

    return createdPosts, nil
}

func (uc *ThreadUseCase) GetThreadPosts(slugOrID string, limit, since int, sort string, desc bool) (models.Posts, error) {
    thread, err := uc.threadRepo.GetThreadBySlugOrID(slugOrID)
    if err != nil {
        return nil, errors.New("thread not found")
    }

    if limit <= 0 {
        limit = 100
    }
    
    validSorts := map[string]bool{"flat": true, "tree": true, "parent_tree": true}
    if !validSorts[sort] {
        sort = "flat"
    }

    return uc.threadRepo.GetThreadPosts(thread.ID, limit, since, sort, desc)
}

func (uc *ThreadUseCase) GetThreadDetails(slugOrID string) (*models.Thread, error) {
    return uc.threadRepo.GetThreadBySlugOrID(slugOrID)
}

func (uc *ThreadUseCase) UpdateThread(slugOrID string, update models.ThreadUpdate) (*models.Thread, error) {
    thread, err := uc.threadRepo.GetThreadBySlugOrID(slugOrID)
    if err != nil {
        return nil, errors.New("thread not found")
    }
    
    if update.Message != "" {
        thread.Message = update.Message
    }
    if update.Title != "" {
        thread.Title = update.Title
    }
    
    return uc.threadRepo.UpdateThread(thread)
}

func (uc *ThreadUseCase) VoteForThread(slugOrID string, vote models.Vote) (*models.Thread, error) {
    if vote.Voice != -1 && vote.Voice != 1 {
        return nil, errors.New("voice must be -1 or 1")
    }
	
	// Проверяем существование пользователя
    if _, err := uc.userRepo.GetUserByNickname(vote.Nickname); err != nil {
        return nil, errors.New("user not found")
    }
    
    // Получаем ветку
    thread, err := uc.threadRepo.GetThreadBySlugOrID(slugOrID)
    if err != nil {
        return nil, errors.New("thread not found")
    }
    
    // Добавляем/обновляем голос
    err = uc.threadRepo.AddOrUpdateVote(thread.ID, vote)
    if err != nil {
        return nil, err
    }
    
    // Получаем обновленную ветку
    return uc.threadRepo.GetThreadByID(thread.ID)
}