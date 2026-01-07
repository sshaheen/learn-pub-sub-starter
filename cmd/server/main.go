package main

import (
	"fmt"
	"log"

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

	qKey := routing.GameLogSlug + ".*"

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, qKey, pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatal("error binding to peril_topic")
	}

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		} else if input[0] == "pause" {
			fmt.Println("sending a pause message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		} else if input[0] == "resume" {
			fmt.Println("sending a resume message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		} else if input[0] == "quit" {
			fmt.Println("exiting program")
			break
		} else {
			fmt.Println("idk what you mean mang")
		}
	}
}
