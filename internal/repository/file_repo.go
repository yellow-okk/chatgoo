package repository

import (
	"context"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// FileRepository defines data access methods for files.
type FileRepository interface {
	Create(ctx context.Context, f *model.File) error
	GetByID(ctx context.Context, fileID int64) (*model.File, error)
}

type fileRepository struct {
	db container.DB
}

// NewFileRepository creates a FileRepository.
func NewFileRepository(db container.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, f *model.File) error {
	query := `
		INSERT INTO files (uploader_id, file_name, file_url, file_size, file_type, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING file_id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		f.UploaderID, f.FileName, f.FileURL, f.FileSize, f.FileType, f.MimeType,
	).Scan(&f.FileID, &f.CreatedAt)
}

func (r *fileRepository) GetByID(ctx context.Context, fileID int64) (*model.File, error) {
	var f model.File
	query := `
		SELECT file_id, uploader_id, file_name, file_url, file_size, file_type, mime_type, created_at
		FROM files WHERE file_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, fileID).Scan(
		&f.FileID, &f.UploaderID, &f.FileName, &f.FileURL,
		&f.FileSize, &f.FileType, &f.MimeType, &f.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
