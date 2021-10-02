package queue

import (
	"encoding/json"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/streadway/amqp"
)

type QueueManager struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue *amqp.Queue
}

func NewQueueManager(cfg *config.QueueConfig) (*QueueManager, error) {
	conn, err := connectToManager()
	if err != nil {
		return &QueueManager{}, err
	}

	ch, err := connectToChannel(conn)
	if err != nil {
		return &QueueManager{}, err
	}

	queue, err := declareQueue(ch, cfg.QueueName)
	if err != nil {
		return &QueueManager{}, err
	}

	return &QueueManager{
		conn:  conn,
		ch:    ch,
		queue: queue,
	}, nil
}

func (q *QueueManager) SendMessage(message interface{}) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = publishMessage(q.ch, q.queue, jsonMessage)

	return err
}

func (q *QueueManager) CloseConnection() {
	closeConnection(q.conn)
}

func (q *QueueManager) CloseChannel() {
	closeChannel(q.ch)
}
