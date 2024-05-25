package main

import (
	"fmt"
	"os"
	"os/signal"

	gamelogic "github.com/Madalosso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/Madalosso/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/Madalosso/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connection, err := amqp.Dial(routing.BrokerUrl)
	if err != nil {
		fmt.Println(err)
	}
	defer connection.Close()

	fmt.Println("Connection established")

	username, err := gamelogic.ClientWelcome()

	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, 1)

	signalChan := make(chan os.Signal, 1)
	//subscribe?
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("\nConnection shutdown\n")
}
