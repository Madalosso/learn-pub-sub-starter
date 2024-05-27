package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// simpleQueueType 0 = Durable
// simpleQueueType 1 = Transient
func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	clientChan, err := conn.Channel()
	if err != nil {
		fmt.Println("Error while creating connection channel", err)
	}
	durable := simpleQueueType == 0
	autoDelete := simpleQueueType == 1
	exclusive := simpleQueueType == 1

	amqpTable := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := clientChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqpTable)
	if err != nil {
		fmt.Println("Error while declaring new queue", err)
	}
	err = clientChan.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Println("Error on queue binding", err)
	}

	return clientChan, queue, nil
}
