package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/waggertron/twelve-factor-go/internal/domain"
	"go.opentelemetry.io/otel"
)

const (
	QueueName = "orders.v1"
	TaskType  = "complete-order"
)

type Enqueuer interface {
	Enqueue(context.Context, domain.Job) error
}

type Client struct{ client *asynq.Client }

func RedisOptions(redisURL string) (asynq.RedisClientOpt, error) {
	options, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}
	redisOptions, ok := options.(asynq.RedisClientOpt)
	if !ok {
		return asynq.RedisClientOpt{}, fmt.Errorf("redis URL must identify one Redis server")
	}
	return redisOptions, nil
}

func NewClient(redisURL string) (*Client, error) {
	options, err := RedisOptions(redisURL)
	if err != nil {
		return nil, err
	}
	return &Client{client: asynq.NewClient(options)}, nil
}

func (c *Client) Enqueue(ctx context.Context, job domain.Job) error {
	if err := job.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = c.client.EnqueueContext(
		ctx,
		asynq.NewTask(TaskType, payload),
		asynq.Queue(QueueName),
		asynq.MaxRetry(2),
	)
	return err
}

func (c *Client) Close() error { return c.client.Close() }

type Processor struct {
	Orders interface {
		Claim(context.Context, string) (bool, error)
		Complete(context.Context, string) (bool, error)
	}
	OnResult func(context.Context, domain.Job, string)
}

func (p Processor) Handle(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("twelve-factor-orders/worker").Start(ctx, "orders.complete")
	defer span.End()
	var job domain.Job
	if err := json.Unmarshal(task.Payload(), &job); err != nil {
		return fmt.Errorf("decode job: %w", asynq.SkipRetry)
	}
	if err := job.Validate(); err != nil {
		return fmt.Errorf("validate job: %w", asynq.SkipRetry)
	}
	claimed, err := p.Orders.Claim(ctx, job.OrderID)
	if err != nil || !claimed {
		if err == nil && p.OnResult != nil {
			p.OnResult(ctx, job, "skipped")
		}
		return err
	}
	completed, err := p.Orders.Complete(ctx, job.OrderID)
	if err == nil && completed && p.OnResult != nil {
		p.OnResult(ctx, job, "completed")
	}
	return err
}
