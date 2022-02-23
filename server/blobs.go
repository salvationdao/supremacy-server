package server

import "time"

// Blob is a single attachment item on the platform
type Blob struct {
	ID            BlobID     `json:"id" db:"id"`
	FileName      string     `json:"fileName" db:"file_name"`
	MimeType      string     `json:"mimeType" db:"mime_type"`
	FileSizeBytes int64      `json:"fileSizeBytes" db:"file_size_bytes"`
	Extension     string     `json:"extension" db:"extension"`
	File          []byte     `json:"file" db:"file"`
	Views         int        `json:"views" db:"views"`
	Hash          *string    `json:"hash" db:"hash"`
	DeletedAt     *time.Time `json:"deletedAt" db:"deleted_at"`
	UpdateAt      *time.Time `json:"updatedAt" db:"updated_at"`
	CreatedAt     *time.Time `json:"createdAt" db:"created_at"`
}
