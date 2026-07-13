package handlers

import (
	"io"

	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
)

const defaultUploadKind = "campaign_cover"

type FileHandler struct {
	upload *filesuc.UploadService
	orgs   app.OrganizationRepository
}

func NewFileHandler(upload *filesuc.UploadService, orgs app.OrganizationRepository) *FileHandler {
	return &FileHandler{upload: upload, orgs: orgs}
}

// Upload implements Etapa 4 §6: sube primero, asocia después — el form de
// creación de campaña no se bloquea esperando la imagen.
func (h *FileHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	org, err := h.orgs.GetPersonalByUserID(c.Context(), userID)
	if err != nil {
		return dto.WriteError(c, err)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return dto.WriteError(c, err)
	}

	f, err := fileHeader.Open()
	if err != nil {
		return dto.WriteError(c, err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return dto.WriteError(c, err)
	}

	kind := c.FormValue("kind", defaultUploadKind)

	result, err := h.upload.Upload(c.Context(), filesuc.UploadInput{
		OrganizationID: org.ID,
		Kind:           kind,
		MimeType:       fileHeader.Header.Get("Content-Type"),
		Data:           data,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":  result.ID.String(),
		"url": result.URL,
	})
}
