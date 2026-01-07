package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rmq_str := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rmq_str)
	if err != nil {
		log.Fatal("error with dialing rmq")
	}

	defer conn.Close()
	fmt.Println("Connection was successful")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("error creating channel")
	}

	input, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("couldn't get input from client welcome")
	}

	qName := strings.Join([]string{routing.PauseKey, input}, ".")

	gs := gamelogic.NewGameState(input)

	if err := pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, qName, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs)); err != nil {
		log.Fatal("couldn't pause")
	}

	mqName := strings.Join([]string{routing.ArmyMovesPrefix, input}, ".")
	routingKey := strings.Join([]string{routing.ArmyMovesPrefix, "*"}, ".")
	if err := pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, mqName, routingKey, pubsub.SimpleQueueTransient, handlerMove(ch, gs)); err != nil {
		log.Fatal("couldn't connect to move queue")
	}

	if err := pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gs)); err != nil {
		log.Fatal("error with war queue")
	}

ClientLoop:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			gs.CommandSpawn(input)
		case "move":
			mv, err := gs.CommandMove(input)
			if err != nil {
				log.Fatal("could not move")
				continue
			}
			pubsub.PublishJSON(ch, routing.ExchangePerilTopic, strings.Join([]string{routing.ArmyMovesPrefix, mv.Player.Username}, "."), mv)
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break ClientLoop
		default:
			fmt.Println("command unrecognized")
		}
	}
}
