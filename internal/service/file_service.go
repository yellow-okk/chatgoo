package service

import (
	"context"

	"chatgoo/internal/model"
	"chatgoo/internal/repository"
)

// FileService defines file business operations.
type FileService interface {
	GetByID(ctx context.Context, fileID int64) (*model.File, error)
}

type fileService struct {
	fileRepo repository.FileRepository
}

// NewFileService creates a FileService.
func NewFileService(fileRepo repository.FileRepository) FileService {
	return &fileService{fileRepo: fileRepo}
}

func (s *fileService) GetByID(ctx context.Context, fileID int64) (*model.File, error) {
	return s.fileRepo.GetByID(ctx, fileID)
}
