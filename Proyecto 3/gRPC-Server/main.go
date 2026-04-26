package main

import (
	"context"
	"log"
	"net"

	pb "gRPC-Server/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWarReportServiceServer
}

func (s *server) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	log.Printf("gRPC Server recibio reporte: Country=%s, WarplanesInAir=%d, WarplanesInWater=%d", req.Country, req.WarplanesInAir, req.WarplanesInWater)

	// Aca debe publicar o escribir en RabbitMQ

	return &pb.WarReportResponse{Status: "Reporte recibido"}, nil
}

func main() {
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
