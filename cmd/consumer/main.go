package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ernane-sabino/goIntensivo/internal/order/infra/database"
	"github.com/ernane-sabino/goIntensivo/internal/order/usecase"
	"github.com/ernane-sabino/goIntensivo/pkg/rabbitmq"
	_ "github.com/mattn/go-sqlite3" // sqlite3
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	db, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repository := database.NewOrderRepository(db)
	uc := usecase.CalculateFinalPriceUseCase{OrderRepository: repository}

	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	out := make(chan amqp.Delivery) // channel
	go rabbitmq.Consume(ch, out)    // T2

	for msg := range out {
		var inputDTO usecase.OrderInputDTO
		err := json.Unmarshal(msg.Body, &inputDTO)
		if err != nil {
			panic(err)
		}
		outputDTO, err := uc.Execute(inputDTO)
		if err != nil {
			panic(err)
		}
		msg.Ack(false)
		fmt.Println(outputDTO)
		time.Sleep(500 * time.Millisecond)
	}
}
