from locust import HttpUser, task, between
import random, time

class WarTrafic(HttpUser):
    wait_time = between(0.1, 0.5)
    
    @task
    def send_report(self):
        countries = ["USA", "RUS", "CHN", "ESP", "GMT"]
        payload = {
            "country": random.choice(countries),
            "warplanes_in_air": random.randint(0, 50),
            "warplanes_in_water": random.randint(0, 30),
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ")
        }
        self.client.post("/grpc-201602929", json=payload)