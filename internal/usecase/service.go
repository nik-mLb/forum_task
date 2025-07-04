package usecase

import (
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/repository"
)

type ServiceUseCase struct {
	repo repository.IServiceRepo
}

func NewServiceUseCase(repo repository.IServiceRepo) *ServiceUseCase{
	return &ServiceUseCase{
		repo: repo,
	}
}

type IServiceUseCase interface {
	Clear() error
	GetStatus() (*models.Status, error)
}

func (uc *ServiceUseCase) Clear() error{
	if err := uc.repo.Clear(); err != nil {
		return err
	}

	return nil
}

func (uc *ServiceUseCase) GetStatus() (*models.Status, error) {
	status, err := uc.repo.SelectStatus()
	if err != nil {
		return nil, err
	}

	return status, nil
} 