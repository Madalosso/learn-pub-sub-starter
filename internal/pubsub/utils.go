package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	data, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Error on publishJSON: ", err)
	}

	ctx := context.Background()
	fmt.Println("msgData val", val)
	fmt.Println("msgData post", data)
	return ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: data})
}

// Check if ok
type SimpleQueueType int

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Error while declare+bind on subscribeJSON", err)
		return err
	}

	msgs, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)
	go func() {
		defer channel.Close()
		for msg := range msgs {
			var msgValue T //routing.PlayingState
			json.Unmarshal(msg.Body, &msgValue)
			handler(msgValue)
			msg.Ack(false)
		}
	}()
	return nil
}

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

	queue, err := clientChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		fmt.Println("Error while declaring new queue", err)
	}
	err = clientChan.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Println("Error on queue binding", err)
	}

	return clientChan, queue, nil
}
