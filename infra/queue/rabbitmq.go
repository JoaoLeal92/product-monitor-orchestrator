package queue

import (
	"github.com/streadway/amqp"
)

func connectToManager() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	return conn, err
}

func connectToChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()

	return ch, err
}

func declareQueue(ch *amqp.Channel, queueName string) (*amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	return &q, err
}

func publishMessage(ch *amqp.Channel, q *amqp.Queue, msg []byte) error {
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
		})

	return err
}

func closeConnection(conn *amqp.Connection) {
	conn.Close()
}

func closeChannel(ch *amqp.Channel) {
	ch.Close()
}
