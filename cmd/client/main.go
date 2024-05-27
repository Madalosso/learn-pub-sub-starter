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
	if err != nil {
		fmt.Println("Error welcoming user: ", err)
	}
	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, 1)

	if err != nil {
		fmt.Println("Error binding queue: ", err)
	}
	gs := gamelogic.NewGameState(username)

	// declare/bind army_moves queue for this client
	moveChannel, moveQueue, _ := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, "army_moves."+gs.GetUsername(), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), 1)
	_ = moveQueue
	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, gs.GetUsername()), routing.PauseKey, 1, handlerPause(gs))

	// subscribe for army_moves matching any player
	pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, "army_moves."+gs.GetUsername(), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), 1, handlerMove(gs))

repl:
	for {
		input := gamelogic.GetInput()
		if len(input) != 0 {
			command := input[0]
			switch command {
			case "spawn":
				fmt.Println("Spawn cmd. args: ", input[1:])
				gs.CommandSpawn(input)
				//shouldn't this be also communicated?
			case "move":
				fmt.Println("move cmd. args: ", input[1:])
				move, err := gs.CommandMove(input)
				if err != nil {
					fmt.Println("Error while moving: ", err)
				}
				err = pubsub.PublishJSON(moveChannel, routing.ExchangePerilTopic, "army_moves."+gs.GetUsername(), move)
				if err != nil {
					fmt.Println("Error while publishing move: ", err)
				} else {
					fmt.Println("Move successfully published!")
				}
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		if outcome == gamelogic.MoveOutcomeMakeWar {
			// pubsub.PublishJSON()
			// return pubsub.NackRequeue

		}

		if outcome == gamelogic.MoveOutComeSafe || outcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}
