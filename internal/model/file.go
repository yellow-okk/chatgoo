package model

import "time"

// File represents an uploaded file.
type File struct {
	FileID     int64     `json:"file_id" db:"file_id"`
	UploaderID int64     `json:"uploader_id" db:"uploader_id"`
	FileName   string    `json:"file_name" db:"file_name"`
	FileURL    string    `json:"file_url" db:"file_url"`
	FileSize   int64     `json:"file_size" db:"file_size"`
	FileType   string    `json:"file_type" db:"file_type"`
	MimeType   string    `json:"mime_type" db:"mime_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
