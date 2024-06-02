package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

type AckType int

const (
	Ack         AckType = 0
	NackRequeue         = 1
	NackDiscard         = 2
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	// can't we use channel instead of conn since we only use conn to open a new channel?
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Error while declare+bind on subscribeJSON", err)
		return err
	}

	err = channel.Qos(10, 0, false)

	if err != nil {
		fmt.Println("Error on subscribe channel.Qos", err)
		return err
	}

	msgs, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)

	go func() {
		defer channel.Close()
		for msg := range msgs {
			var msgValue T //routing.PlayingState
			json.Unmarshal(msg.Body, &msgValue)
			ackType := handler(msgValue)
			switch ackType {
			case Ack:
				fmt.Println("msg.Ack(false)")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("msg.Nack(false, true)")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("msg.Nack(false, false)")
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	// can't we use channel instead of conn since we only use conn to open a new channel?
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Error while declare+bind on subscribeJSON", err)
		return err
	}

	msgs, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)
	go func() {
		defer channel.Close()
		for msg := range msgs {
			var buffer bytes.Buffer
			buffer.Write(msg.Body)
			var msgValue T //routing.PlayingState
			decoder := gob.NewDecoder(&buffer)
			decoder.Decode(&msgValue)
			handler(msgValue)
			ackType := handler(msgValue)
			switch ackType {
			case Ack:
				msg.Ack(false)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
