package rtask

import (
	"time"

	"github.com/hibiken/asynq"
	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

type Client struct {
	asyncClient *asynq.Client
	logger      *slog.Logger
}

func New(redisDsnURI string, logger *slog.Logger) (Client, error) {
	const op = "create rtask client"

	opt, err := asynq.ParseRedisURI(redisDsnURI)
	if err != nil {
		return Client{}, e.Wrap(op, err)
	}

	return Client{
		asyncClient: asynq.NewClient(opt),
		logger:      logger.With("from", "rtask client"),
	}, nil
}

func (c *Client) Enqueue(task *asynq.Task, timeout time.Duration) error {
	const op = "task client: enqueue"

	info, err := c.asyncClient.Enqueue(task, asynq.Timeout(timeout))
	if err != nil {
		return e.Wrap(op, err)
	}
	c.logger.Debug("successfully enqueued task",
		slog.String("task id", info.ID),
		slog.String("task type", info.Type))
	return nil
}
