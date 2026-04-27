package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type WarReport struct {
	Country          string `json:"country"`
	WarplanesInAir   int    `json:"warplanes_in_air"`
	WarplanesInWater int    `json:"warplanes_in_water"`
	Timestamp        string `json:"timestamp"`
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Conectar a RabbitMQ
	url := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal: %v", err)
	}
	defer ch.Close()

	queueName := "war_reports"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error al declarar la cola: %v", err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error al consumir mensajes: %v", err)
	}

	log.Println("Consumer iniciado. Esperando mensajes...")

	// Crea cliente de Valkey
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	rdb.FlushDB(ctx)

	for msg := range msgs {
		var report WarReport
		err := json.Unmarshal(msg.Body, &report)
		if err != nil {
			log.Printf("Error al deserializar mensaje: %v", err)
			msg.Nack(false, false)
			continue
		}

		log.Printf("Reporte recibido: Country=%s, WarplanesInAir=%d, WarplanesInWater=%d", report.Country, report.WarplanesInAir, report.WarplanesInWater)

		// Aca guardar en Valkey
		rdb.ZAdd(ctx, "warplanes_air_top", redis.Z{
			Score:  float64(report.WarplanesInAir),
			Member: report.Country,
		})

		rdb.ZAdd(ctx, "warplanes_water_top", redis.Z{
			Score:  float64(report.WarplanesInWater),
			Member: report.Country,
		})

		rdb.HIncrBy(ctx, "warplanes_air_histogram", fmt.Sprintf("%d", report.WarplanesInAir), 1)
		rdb.HIncrBy(ctx, "warplanes_water_histogram", fmt.Sprintf("%d", report.WarplanesInWater), 1)

		if report.Country == "GTM" {
			rdb.ZAdd(ctx, "warplanes_air_series:GTM", redis.Z{
				Score:  float64(report.WarplanesInAir),
				Member: report.Timestamp,
			})
			rdb.ZAdd(ctx, "warplanes_water_series:GTM", redis.Z{
				Score:  float64(report.WarplanesInWater),
				Member: report.Timestamp,
			})

			rdb.Incr(ctx, "reports_count:GTM")
		}

		msg.Ack(false)
	}
}
