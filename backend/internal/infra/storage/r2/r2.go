// Package r2 implements app.FileStorage sobre Cloudflare R2 (S3-compatible):
// mismo contrato que internal/infra/storage/local, pero persistente entre
// deploys — el disco local del contenedor de Render se borra en cada
// redeploy/reinicio (STORAGE_DIR apuntaba a /tmp, que es efímero por
// definición), así que cualquier imagen subida en producción desaparecía
// tarde o temprano.
package r2

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// Config carries what's needed to talk to R2: accountID arma el endpoint
// S3-compatible (https://<accountID>.r2.cloudflarestorage.com), y publicURL
// es el dominio público del bucket (r2.dev o un dominio custom conectado) —
// distinto del endpoint de la API, que no sirve archivos directamente.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string
}

func New(cfg Config) *Storage {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "auto", // R2 no tiene regiones reales; el SDK igual lo exige.
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})

	return &Storage{client: client, bucket: cfg.Bucket, publicURL: cfg.PublicURL}
}

func (s *Storage) Save(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(http.DetectContentType(data)),
	})
	if err != nil {
		return fmt.Errorf("r2: put object: %w", err)
	}
	return nil
}

func (s *Storage) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, key)
}
