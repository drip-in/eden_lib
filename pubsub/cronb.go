package pubsub

import (
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type AsyncScheduler struct {
	scheduler *asynq.Scheduler
}

func NewAsyncScheduler(cfg AsyncQueueConfig) AsyncScheduler {
	redisCfg := asynq.RedisClientOpt{Addr: cfg.RedisURL}
	scheduler := asynq.NewScheduler(redisCfg, &asynq.SchedulerOpts{
		PreEnqueueFunc: func(task *asynq.Task, opts []asynq.Option) {
			if task != nil {
				loc, err := time.LoadLocation(cfg.TimeZone)
				if err != nil {
					log.Printf("error while before enqueuing task %s,err:%s", task.Type(), err.Error())
				}
				now := time.Now()
				timeNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, loc)
				taskId := timeNow.Format(time.RFC3339) + task.Type()
				opts = append(opts, asynq.TaskID(taskId))
				*task = *asynq.NewTask(
					task.Type(),
					task.Payload(),
					opts...,
				)
			}
		},
		PostEnqueueFunc: func(task *asynq.TaskInfo, err error) {
			if err != nil {
				var taskId string
				if task != nil {
					taskId = task.ID
				}
				log.Printf("error while enqueuing task %s,err:%s", taskId, err.Error())
			}
		},
	})
	return AsyncScheduler{
		scheduler: scheduler,
	}
}

func (app *AsyncScheduler) Run() error {
	return app.scheduler.Start()
}

func (app *AsyncScheduler) Register(cronSpec string, cronJob Job[JobArg]) {
	if cronJob.Config().Retention <= 0 {
		log.Fatalf("CronJob must have retention %s \n", cronJob.Name())
		return
	}
	task, err := newAsyncTask(cronJob)
	if err != nil {
		log.Fatalf("CronJob NewAsynqTask %s failed Error: %v\n", cronJob.Name(), err)
		return
	}
	if _, err = app.scheduler.Register(cronSpec, task, newAsynqOptions(cronJob.Config())...); err != nil {
		log.Fatalf("registerCronJob %s failed Error: %v\n", cronJob.Name(), err)
	}
}

func newAsyncTask(job Job[JobArg]) (*asynq.Task, error) {
	payload, err := job.Arg().JSON()
	if err != nil {
		return nil, err
	}
	opts := newAsynqOptions(job.Config())
	task := asynq.NewTask(job.Name(), payload, opts...)
	return task, err
}
