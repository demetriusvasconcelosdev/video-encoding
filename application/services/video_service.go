package services

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/dl/encoding-video/application/repositories"
	"github.com/dl/encoding-video/domain"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	r, err := client.Bucket(bucketName).Object(v.Video.FilePath).NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := os.Create(filepath.Join(os.Getenv("localStoragePath"), v.Video.ID+".mp4"))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}
