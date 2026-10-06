package domain

import (
	"errors"
	"time"
)

type Video struct {
	ID         string
	ResourceID string
	FilePath   string
	CreatedAt  time.Time
}

func NewVideo() *Video {
	return &Video{}
}

func (v *Video) Validate() error {
	if v.ID == "" {
		return errors.New("id is required")
	}
	if v.ResourceID == "" {
		return errors.New("resource id is required")
	}
	if v.FilePath == "" {
		return errors.New("file path is required")
	}
	return nil
}
