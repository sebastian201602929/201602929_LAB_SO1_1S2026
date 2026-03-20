# Descarga el modulo de kernel
sudo rmmod modulo

# Eliminacion del cronjob y detener su ejecución
crontab -r | crontab -

# Elimina contenedor de Grafana
cd /home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/Grafana
docker-compose down