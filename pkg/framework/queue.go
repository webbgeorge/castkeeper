package framework

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	visibilityTimeout = time.Minute * 30
	maxReceives       = 5
	backoffInterval   = time.Second * 10
	backoffExponent   = 2
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type QueueTask struct {
	gorm.Model
	QueueName    string `validate:"required,gte=1"`
	VisibleAfter time.Time
	ReceiveCount uint
	Data         any `gorm:"serializer:json"`
}

func (t *QueueTask) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(t)
	if err != nil {
		return fmt.Errorf("queue task not valid: %w", err)
	}
	return nil
}

func PushQueueTask(ctx context.Context, db *gorm.DB, queueName string, data any) error {
	queueTask := QueueTask{
		QueueName:    queueName,
		VisibleAfter: time.Now(),
		ReceiveCount: 0,
		Data:         data,
	}
	if err := db.Create(&queueTask).Error; err != nil {
		return err
	}
	return nil
}

func PopQueueTask(ctx context.Context, db *gorm.DB, queueName string) (QueueTask, error) {
	var queueTask QueueTask
	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where(
				"queue_name = ? AND visible_after < ? AND receive_count <= ?",
				queueName,
				time.Now(),
				maxReceives,
			).
			Order("created_at asc").
			First(&queueTask)
		if result.Error != nil {
			return result.Error
		}

		queueTask.VisibleAfter = time.Now().Add(visibilityTimeout)
		queueTask.ReceiveCount++

		result = tx.Save(&queueTask)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})
	if err != nil {
		return queueTask, err
	}

	return queueTask, nil
}

func completeQueueTask(db *gorm.DB, queueTask QueueTask) error {
	if err := db.Delete(&queueTask).Error; err != nil {
		return err
	}
	return nil
}

func returnQueueTask(db *gorm.DB, queueTask QueueTask) error {
	backoffFactor := math.Pow(backoffExponent, float64(queueTask.ReceiveCount))
	timeUntilNextTry := backoffInterval * time.Duration(backoffFactor)
	queueTask.VisibleAfter = time.Now().Add(timeUntilNextTry)
	if err := db.Save(&queueTask).Error; err != nil {
		return err
	}
	return nil
}

type QueueWorker struct {
	DB        *gorm.DB
	QueueName string
	HandlerFn func(ctx context.Context, data any) error
}

func (w *QueueWorker) Start(ctx context.Context) error {
	for {
		// handle cancellation at top to ensure it runs on every iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		qt, err := PopQueueTask(ctx, w.DB, w.QueueName)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("failed to pop task from queue '%s' with err '%s'", w.QueueName, err.Error()))
			}
			time.Sleep(10 * time.Second) // no jobs on queue, or other error, wait before next poll
			continue
		}

		err = w.HandlerFn(ctx, qt.Data)
		if err != nil {
			GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("failed to process task '%d' of queue '%s' with err '%s'", qt.ID, w.QueueName, err.Error()))
			err = returnQueueTask(w.DB, qt)
			if err != nil {
				GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("failed to return task '%d' to queue '%s' with err '%s'", qt.ID, w.QueueName, err.Error()))
			}
			continue
		}

		err = completeQueueTask(w.DB, qt)
		if err != nil {
			GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("failed to mark task '%d' complete from queue '%s' with err '%s'", qt.ID, w.QueueName, err.Error()))
		}

		GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("successfully processed task '%d' from queue '%s'", qt.ID, w.QueueName))
	}
}
