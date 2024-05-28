package main

import (
	"fmt"
	"time"

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

	publishCh, err := connection.Channel()
	if err != nil {
		fmt.Println("Couldnt' get channel from connection: ", err)
	}
	// _, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, 1)

	gs := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, gs.GetUsername()), routing.PauseKey, 1, handlerPause(gs))
	pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix), 0, handlerWar(gs, publishCh))
	pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, "army_moves."+gs.GetUsername(), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), 1, handlerMove(gs, publishCh))

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
				err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, "army_moves."+gs.GetUsername(), move)
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

func handlerMove(gs *gamelogic.GameState, warChannel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		if outcome == gamelogic.MoveOutcomeMakeWar {
			// what should be the msg content here?
			warMove := gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: gs.Player}
			err := pubsub.PublishJSON(warChannel, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()), warMove)
			if err != nil {
				fmt.Println("Issue publishing war recognition: \n Requeuing the msg", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		if outcome == gamelogic.MoveOutComeSafe {
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, logChannel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(warMove gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(warMove)
		fmt.Printf("Handler war: %v, outcome: %v, winner: %v, loser: %v\n", warMove, outcome, winner, loser)
		var returnValue pubsub.AckType
		var gameLog routing.GameLog
		gameLog.CurrentTime = time.Now()
		gameLog.Username = gs.GetUsername()
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			returnValue = pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			returnValue = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			gameLog.Message = fmt.Sprintf("%s won a war against %s", winner, loser)
			returnValue = pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			gameLog.Message = fmt.Sprintf("%s won a war against %s", winner, loser)
			returnValue = pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			gameLog.Message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			returnValue = pubsub.Ack
		default:
			fmt.Println("Handler war: Outcome doesn't match expected values.", outcome)
			returnValue = pubsub.NackDiscard
		}
		if len(gameLog.Message) > 0 {
			logErr := pubsub.PublishGob(logChannel, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()), gameLog)
			if logErr != nil {
				fmt.Println("Erro publishing game log via PublishGob: ", logErr)
				return pubsub.NackRequeue
			}
		}

		return returnValue
	}
}
