package main

import (
	"fmt"

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

	gs := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, 1, handlerPause(gs))

repl:
	for {
		input := gamelogic.GetInput()
		if len(input) != 0 {
			command := input[0]
			switch command {
			case "spawn":
				fmt.Println("Spawn cmd. args: ", input[1:])
				gs.CommandSpawn(input)
			case "move":
				fmt.Println("move cmd. args: ", input[1:])
				gs.CommandMove(input)
			case "status":
				gs.CommandStatus()
			case "help":
				gamelogic.PrintClientHelp()
			case "spam":
				fmt.Println("Spamming not allowed yet!")
			case "quit":
				gamelogic.PrintQuit()
				break repl
			default:
				fmt.Printf("The command `%s` isn't valid\n", command)
			}
		}
	}

	// signalChan := make(chan os.Signal, 1)
	// //subscribe?
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	fmt.Printf("\nConnection shutdown\n")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
