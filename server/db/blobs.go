package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// Blob returns a blob by given ID
func FindBlob(ctx context.Context, conn Conn, result *server.Blob, blobID server.BlobID) error {
	q := `
		SELECT id, file_name, file_size_bytes, extension, mime_type, file, views
		FROM blobs
		WHERE id = $1`
	err := pgxscan.Get(ctx, conn, result, q, blobID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// BlobInsert inserts a new blob
func BlobInsert(ctx context.Context, conn Conn, result *server.Blob, id server.BlobID, fileName string, mimeType string, fileSizeBytes int64, extension string, file []byte, hash *string) error {
	q := `
		INSERT INTO blobs (id, file_name, mime_type, file_size_bytes, extension, file, hash) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET file_name = EXCLUDED.file_name,
			mime_type = EXCLUDED.mime_type,
			file_size_bytes = EXCLUDED.file_size_bytes,
			extension = EXCLUDED.extension,
			file = EXCLUDED.file,
			hash = EXCLUDED.hash
		RETURNING id, file_name, mime_type, file_size_bytes, extension, file, hash`
	err := pgxscan.Get(ctx, conn, result, q, id, fileName, mimeType, fileSizeBytes, extension, file, hash)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
