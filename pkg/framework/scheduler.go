package framework

import (
	"context"
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

func (r *TaskScheduler) processTask(ctx context.Context, taskDef ScheduledTaskDefinition) {
	r.DB.Transaction(func(tx *gorm.DB) error {
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
