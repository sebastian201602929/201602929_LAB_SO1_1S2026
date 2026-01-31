package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	VM        int       `json:"vm"`
	Carnet    int       `json:"carnet"`
}

type APIResponse struct {
	APIname    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     int    `json:"carnet"`
}

var vm1_url string = ""
var vm2_url string = ""
var api1_port string = ""
var api2_port string = ""
var api3_port string = ""

func main() {
	vm1_url = os.Getenv("VM1_URL")
	vm2_url = os.Getenv("VM2_URL")
	api1_port = os.Getenv("API1_PORT")
	api2_port = os.Getenv("API2_PORT")
	api3_port = os.Getenv("API3_PORT")

	// Valores por defecto si las variables de entorno no están definidas
	if api1_port == "" {
		api1_port = "8081"
	}
	if api2_port == "" {
		api2_port = "8082"
	}
	if api3_port == "" {
		api3_port = "8083"
	}
	if vm1_url == "" {
		vm1_url = "localhost"
	}
	if vm2_url == "" {
		vm2_url = "localhost"
	}

	// Habilita endpoints
	http.HandleFunc("/health", healthHandler())
	http.HandleFunc("/api1/201602929/call-api2", callAPI2Handler())
	http.HandleFunc("/api1/201602929/call-api3", callAPI3Handler())

	fmt.Println("Iniciando API1 en el puerto :", api1_port)

	if err := http.ListenAndServe(":"+api1_port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Construye objeto de respuesta
		response := HealthResponse{
			Status:    "UP",
			Message:   "API1 is Ready",
			Timestamp: time.Now(),
			VM:        1,
			Carnet:    201602929,
		}

		// Código de estado OK
		w.WriteHeader(http.StatusOK)

		// Respuesta en formato JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func callAPI2Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Get("http://" + vm1_url + ":" + api2_port + "/health")

		// Respuesta NO OK
		if err != nil || resp.StatusCode != http.StatusOK {
			response := APIResponse{
				APIname:    "API2",
				Message:    "ERROR: The API2 located on the VM1 is not working",
				Connection: false,
				Carnet:     201602929,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

			return
		}

		defer resp.Body.Close()

		// Respuesta OK
		response := APIResponse{
			APIname:    "API2",
			Message:    "The API2 located on the VM1 is working",
			Connection: true,
			Carnet:     201602929,
		}

		// Respuesta en formato JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func callAPI3Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Get("http://" + vm2_url + ":" + api3_port + "/health")

		// Respuesta NO OK
		if err != nil || resp.StatusCode != http.StatusOK {
			response := APIResponse{
				APIname:    "API3",
				Message:    "ERROR: The API3 located on the VM2 is not working",
				Connection: false,
				Carnet:     201602929,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

			return
		}

		defer resp.Body.Close()

		// Respuesta OK
		response := APIResponse{
			APIname:    "API3",
			Message:    "The API3 located on the VM2 is working",
			Connection: true,
			Carnet:     201602929,
		}

		// Respuesta en formato JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
