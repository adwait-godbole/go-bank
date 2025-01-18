package worker

import (
	"context"

	db "github.com/adwait-godbole/go-bank/db/sqlc"
	"github.com/hibiken/asynq"
)

const (
	DefaultQueue  = "default"
	CriticalQueue = "critical"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerificationEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	return &RedisTaskProcessor{
		server: asynq.NewServer(
			redisOpt,
			asynq.Config{
				Queues: map[string]int{
					DefaultQueue:  5,
					CriticalQueue: 10,
				},
			},
		),
		store: store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerificationEmail, processor.ProcessTaskSendVerificationEmail)

	return processor.server.Start(mux)
}
