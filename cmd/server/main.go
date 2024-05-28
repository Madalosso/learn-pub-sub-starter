package main

import (
	"fmt"
	"os"
	"os/signal"

	gamelogic "github.com/Madalosso/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/Madalosso/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/Madalosso/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connection, err := amqp.Dial(routing.BrokerUrl)
	if err != nil {
		fmt.Println(err)
	}
	defer connection.Close()

	fmt.Println("Connection established")

	serverChan, err := connection.Channel()
	if err != nil {
		fmt.Println("Error while creating connection channel", err)
	}
	err = pubsub.PublishJSON(serverChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	if err != nil {
		fmt.Println("Error while publishing game state to pubsub", err)
	}
	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, fmt.Sprintf("%s.*", routing.GameLogSlug), 0)

	gamelogic.PrintServerHelp()
repl:
	for {
		input := gamelogic.GetInput()
		if len(input) != 0 {
			command := input[0]
			switch command {
			case "pause":
				fmt.Println("The game is being paused")
				err = pubsub.PublishJSON(
					serverChan,
					routing.ExchangePerilDirect,
					routing.PauseKey,
					routing.PlayingState{IsPaused: true},
				)
				if err != nil {
					fmt.Println("Error while publishing game state to pubsub", err)
				}
			case "resume":
				fmt.Println("The game is being resumed")
				err = pubsub.PublishJSON(serverChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
				if err != nil {
					fmt.Println("Error while publishing game state to pubsub", err)
				}
			case "quit":
				fmt.Println("Exiting the game")
				break repl
			default:
				fmt.Printf("The command `%s` isn't valid\n", command)
			}
		}
	}

	signalChan := make(chan os.Signal, 1)
	//subscribe?
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("\nConnection shutdown\n")
}
