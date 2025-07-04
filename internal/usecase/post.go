package usecase

import (
	"errors"
	"nik-mLb/forum_task.com/internal/repository"
	"nik-mLb/forum_task.com/internal/models"
)

type PostUseCase struct {
	postRepo  repository.IPostRepository
	userRepo  repository.IUserRepository
	forumRepo repository.IForumRepository
	threadRepo repository.IThreadRepository
}

func NewPostUseCase(
	postRepo repository.IPostRepository,
	userRepo repository.IUserRepository,
	forumRepo repository.IForumRepository,
	threadRepo repository.IThreadRepository,
) *PostUseCase {
	return &PostUseCase{
		postRepo:  postRepo,
		userRepo:  userRepo,
		forumRepo: forumRepo,
		threadRepo: threadRepo,
	}
}

type IPostUseCase interface {
	GetPostDetails(id int, related []string) (*models.PostFull, error)
	UpdatePost(id int, update models.PostUpdate) (*models.Post, error)
}

func (uc *PostUseCase) GetPostDetails(id int, related []string) (*models.PostFull, error) {
	post, err := uc.postRepo.GetPostByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	postFull := &models.PostFull{Post: post}

	for _, rel := range related {
		switch rel {
		case "user":
			author, err := uc.userRepo.GetUserByNickname(post.Author)
			if err != nil {
				return nil, err
			}
			postFull.Author = author
		case "forum":
			forum, err := uc.forumRepo.GetForumBySlug(post.Forum)
			if err != nil {
				return nil, err
			}
			postFull.Forum = forum
		case "thread":
			thread, err := uc.threadRepo.GetThreadByID(post.Thread)
			if err != nil {
				return nil, err
			}
			postFull.Thread = thread
		}
	}

	return postFull, nil
}

func (uc *PostUseCase) UpdatePost(id int, update models.PostUpdate) (*models.Post, error) {
	post, err := uc.postRepo.GetPostByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	if update.Message != "" && update.Message != post.Message {
		post.Message = update.Message
		post.IsEdited = true
		return uc.postRepo.UpdatePost(post)
	}

	return post, nil
}