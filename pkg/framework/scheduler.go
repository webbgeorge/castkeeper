package framework

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const pollFrequency = time.Minute

type ScheduledTaskState struct {
	TaskName    string `gorm:"primaryKey" validate:"required,gte=1,lte=100"`
	LastRunTime time.Time
}

func (t *ScheduledTaskState) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(t)
	if err != nil {
		return fmt.Errorf("scheduled task not valid: %w", err)
	}
	return nil
}

func getScheduledTask(db *gorm.DB, taskName string) (ScheduledTaskState, error) {
	var task ScheduledTaskState
	result := db.
		Where("task_name = ?", taskName).
		First(&task)
	if result.Error != nil {
		return task, result.Error
	}

	return task, nil
}

// queues scheduled tasks on recurring basis
type TaskScheduler struct {
	DB    *gorm.DB
	Tasks []ScheduledTaskDefinition
}

type ScheduledTaskDefinition struct {
	TaskName string
	Interval time.Duration
}

func (r *TaskScheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(pollFrequency)
	defer ticker.Stop()

	err := r.setupState()
	if err != nil {
		return err
	}

	for {
		for _, td := range r.Tasks {
			r.processTask(ctx, td)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *TaskScheduler) setupState() error {
	for _, td := range r.Tasks {
		err := r.DB.Transaction(func(tx *gorm.DB) error {
			sts := ScheduledTaskState{}
			result := tx.Where("task_name = ?", td.TaskName).First(&sts)
			if result.Error == nil {
				// no err == it already exists, continue
				return nil
			}
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// error other than not found
				return result.Error
			}

			s := ScheduledTaskState{
				TaskName:    td.TaskName,
				LastRunTime: time.Time{}, // 0001-01-01T00:00:00 - i.e. has never run
			}
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskScheduler) processTask(ctx context.Context, taskDef ScheduledTaskDefinition) {
	_ = r.DB.Transaction(func(tx *gorm.DB) error {
		task, err := getScheduledTask(tx, taskDef.TaskName)
		if err != nil {
			GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("failed to get scheduled task '%s': %s", taskDef.TaskName, err.Error()))
			return err
		}

		if time.Since(task.LastRunTime) < taskDef.Interval {
			GetLogger(ctx).DebugContext(ctx, fmt.Sprintf("skipping scheduled task '%s', not enough time has passed", taskDef.TaskName))
			return nil
		}

		err = PushQueueTask(ctx, tx, taskDef.TaskName, "")
		if err != nil {
			GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("failed to get scheduled task '%s': %s", taskDef.TaskName, err.Error()))
			return err
		}

		task.LastRunTime = time.Now()

		if err := tx.Save(&task).Error; err != nil {
			GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("failed to update scheduled task LastRunTime '%s': %s", taskDef.TaskName, err.Error()))
			return err
		}

		return nil
	})
}
