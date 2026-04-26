package main

import (
	"encoding/json"
	"log"
	"net/http"

	pb "API-gRPC-Client/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WarReport struct {
	Country          string `json:"country"`
	WarplanesInAir   int32  `json:"warplanes_in_air"`
	WarplanesInWater int32  `json:"warplanes_in_water"`
	Timestamp        string `json:"timestamp"`
}

func main() {
	// target := "go-server-svc.mumnk8s.svc.cluster.local:50051" # Esto va con el que se crea en los yaml de kubernetes
	target := "localhost:50051"

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar a gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewWarReportServiceClient(conn)
	log.Println("Cliente Go conectado a gRPC. Iniciando envio automatico")

	// Endpoint health
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Recibir reporte desde Rust y enviarlo a gRPC
	http.HandleFunc("/grpc-201602929", func(w http.ResponseWriter, r *http.Request) {
		var report WarReport
		err := json.NewDecoder(r.Body).Decode(&report)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Reporte recibido en Go Deploy 1: %+v", report)

		// Enviar reporte a gRPC
		req := &pb.WarReportRequest{
			Country:          report.Country,
			WarplanesInAir:   report.WarplanesInAir,
			WarplanesInWater: report.WarplanesInWater,
			Timestamp:        report.Timestamp,
		}

		_, err = client.SendReport(r.Context(), req)

		if err != nil {
			log.Printf("Error al enviar reporte a gRPC: %v", err)
			http.Error(w, "Error al enviar reporte a gRPC", http.StatusInternalServerError)
			return
		}

		//log.Printf("Respuesta de gRPC: %s", res.GetStatus())

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Servidor HTTP escuchando en el puerto 8081")
	http.ListenAndServe(":8081", nil)
}
