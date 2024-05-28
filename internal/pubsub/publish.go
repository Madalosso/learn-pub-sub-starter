package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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
	fmt.Println("[JSON] - Publishing value ", val)
	return ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: data})
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var encodedGob bytes.Buffer

	enc := gob.NewEncoder(&encodedGob)
	err := enc.Encode(val)
	if err != nil {
		fmt.Println("Error on publishGob: ", err)
		return err
	}

	ctx := context.Background()
	fmt.Println("[GOB] - Publishing value ", val)
	fmt.Println("[GOB] - Encoded: ", encodedGob.Bytes())
	return ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{ContentType: "application/gob", Body: encodedGob.Bytes()})
}
