package repositories

import (
	"fmt"

	"github.com/dl/encoding-video/domain"
	"github.com/jinzhu/gorm"
)

type JobRepository interface {
	Insert(job *domain.Job) (*domain.Job, error)
	Find(id string) (*domain.Job, error)
	Update(job *domain.Job) (*domain.Job, error)
}

type JobRepositoryDb struct {
	Db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepositoryDb {
	return &JobRepositoryDb{Db: db}
}

func (repo JobRepositoryDb) Insert(job *domain.Job) (*domain.Job, error) {
	err := repo.Db.Create(job).Error
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (repo JobRepositoryDb) Find(id string) (*domain.Job, error) {
	var job domain.Job

	err := repo.Db.Preload("Video").First(&job, "id = ?", id).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("job does not exist")
	}
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (repo JobRepositoryDb) Update(job *domain.Job) (*domain.Job, error) {
	err := repo.Db.Save(job).Error
	if err != nil {
		return nil, err
	}

	return job, nil
}
