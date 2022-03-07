package server

import "time"

// Blob is a single attachment item on the platform
type Blob struct {
	ID            BlobID     `json:"id" db:"id"`
	FileName      string     `json:"file_name" db:"file_name"`
	MimeType      string     `json:"mime_type" db:"mime_type"`
	FileSizeBytes int64      `json:"file_size_bytes" db:"file_size_bytes"`
	Extension     string     `json:"extension" db:"extension"`
	File          []byte     `json:"file" db:"file"`
	Views         int        `json:"views" db:"views"`
	Hash          *string    `json:"hash" db:"hash"`
	DeletedAt     *time.Time `json:"deleted_at" db:"deleted_at"`
	UpdateAt      *time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
}
