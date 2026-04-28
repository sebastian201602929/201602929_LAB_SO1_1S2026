use axum::{
    routing::{get, post},
    Json, Router,
};
use serde::Deserialize;
use tokio::net::TcpListener;

#[derive(Deserialize, serde::Serialize, Debug)]
struct Report {
    country: String,
    warplanes_in_air: i32,
    warplanes_in_water: i32,
    timestamp: String,
}

// Helth check endpoint, necesario para el Google Cloud Load Balancer
async fn health_check() -> &'static str {
    "OK"
}

// Ruta que recibe de Locust
async fn handle_report(Json(payload): Json<Report>) -> String {
    // Simulador de carga para que el HPA se active en clase
    let mut _dummy: u64 = 0;
    for i in 0..5_000_000 {
        _dummy = _dummy.wrapping_add(i);
    }
    
    println!("Recibido reporte de: {}", payload.country);
    //format!("Reporte de {} recibido en GKE", payload.country);

    let grpc_client_url = std::env::var("GRPC_CLIENT_URL").unwrap_or_else(|_| "http://localhost:8081/grpc-201602929".to_string());

    // Aca publicar a API gRPC Client, es otra API REST
    let client = reqwest::Client::new();
    let res = client.post(&grpc_client_url)
        .json(&payload)
        .send()
        .await;

    match res {
        Ok(response) => {
            if response.status().is_success() {
                println!("Reporte enviado al API gRPC Client exitosamente");
                "Status: ok".to_string()
            } else {
                println!("Error al enviar reporte al API gRPC Client: {}", response.status());
                "Status: error".to_string()
            }
        }
        Err(error) => {
            println!("Error al enviar reporte al API gRPC Client: {}", error);
            "Error".to_string()
        }
    }
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/", get(health_check))
        .route("/grpc-201602929", post(handle_report));

    let listener = TcpListener::bind("0.0.0.0:8080").await.unwrap();
    println!("API Rust corriendo en puerto 8080...");
    axum::serve(listener, app).await.unwrap();
}
