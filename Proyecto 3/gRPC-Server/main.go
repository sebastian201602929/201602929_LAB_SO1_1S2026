package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	pb "gRPC-Server/proto"

	"google.golang.org/grpc"

	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	pb.UnimplementedWarReportServiceServer
}

var conn *amqp.Connection
var ch *amqp.Channel
var queueName = "war_reports"

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func (s *server) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	log.Printf("gRPC Server recibio reporte: Country=%s, WarplanesInAir=%d, WarplanesInWater=%d", req.Country, req.WarplanesInAir, req.WarplanesInWater)

	// Aca debe publicar o escribir en RabbitMQ
	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error al convertir el reporte a JSON: %v", err)
		return &pb.WarReportResponse{Status: "Error al procesar el reporte"}, nil
	}

	err = ch.PublishWithContext(context.Background(), "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Printf("Error al publicar el mensaje en RabbitMQ: %v", err)
		return &pb.WarReportResponse{Status: "Error al enviar el reporte"}, nil
	}

	return &pb.WarReportResponse{Status: "Reporte recibido"}, nil
}

func main() {
	// Configurar RabbitMQ
	url := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err = conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal: %v", err)
	}
	defer ch.Close()

	queueName := "war_reports"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error al declarar la cola: %v", err)
	}

	log.Printf("Servidor gRPC conectado a RabbitMQ en %s y listo para publicar mensajes en la cola '%s'", url, queueName)

	// Iniciar servidor gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}

	// Crear servidor gRPC
	s := grpc.NewServer()
	pb.RegisterWarReportServiceServer(s, &server{})

	log.Println("Servidor gRPC escuchando en :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar servidor gRPC: %v", err)
	}
}
