package downloadworker

import (
	"context"

	"gorm.io/gorm"
)

type DownloadWorker struct {
	DB *gorm.DB
}

func (w *DownloadWorker) Start(ctx context.Context) error {
	// TODO fetch undownloaded episodes from the DB
	return nil
}

func (w *DownloadWorker) ProcessDownload() error {
	// TODO download file, save in object storage
	return nil
}
