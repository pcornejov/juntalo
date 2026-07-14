// Package files implements the file upload use case (Etapa 4 §6): sube primero,
// asocia después — el flujo de creación de campaña no se bloquea por la imagen.
package files

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

const maxSizeBytes = 5 * 1024 * 1024 // 5MB

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var (
	errTooLarge        = apperr.New("file_too_large", "El archivo no puede superar los 5MB")
	errUnsupportedType = apperr.New("unsupported_file_type", "Solo se permiten imágenes JPEG, PNG o WEBP")
)

type UploadService struct {
	storage app.FileStorage
	repo    app.FileRepository
}

func NewUploadService(storage app.FileStorage, repo app.FileRepository) *UploadService {
	return &UploadService{storage: storage, repo: repo}
}

type UploadInput struct {
	OrganizationID uuid.UUID
	Kind           string
	MimeType       string
	Data           []byte
}

type UploadResult struct {
	ID  uuid.UUID
	URL string
}

func (s *UploadService) Upload(ctx context.Context, in UploadInput) (UploadResult, error) {
	if len(in.Data) > maxSizeBytes {
		return UploadResult{}, errTooLarge
	}

	// El Content-Type que manda el cliente en el multipart es solo una
	// declaración, no una garantía (auditoría de seguridad: un archivo HTML/SVG
	// disfrazado de "image/png" pasaba la validación anterior). Se detecta el
	// tipo real a partir de los primeros bytes y ESE es el que se valida y se
	// guarda — el header del cliente se ignora por completo para este chequeo.
	detected := http.DetectContentType(in.Data)
	if !allowedMimeTypes[detected] {
		return UploadResult{}, errUnsupportedType
	}

	key := fmt.Sprintf("%s/%s", in.OrganizationID, uuid.New().String())
	if err := s.storage.Save(ctx, key, in.Data); err != nil {
		return UploadResult{}, fmt.Errorf("upload: save: %w", err)
	}

	id, err := s.repo.Create(ctx, app.CreateFileInput{
		OrganizationID: in.OrganizationID,
		Kind:           in.Kind,
		StorageKey:     key,
		MimeType:       detected,
		SizeBytes:      int64(len(in.Data)),
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("upload: create record: %w", err)
	}

	return UploadResult{ID: id, URL: s.storage.PublicURL(key)}, nil
}
