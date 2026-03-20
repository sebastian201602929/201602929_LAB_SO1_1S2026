# Crea contenedor de Grafana
cd /home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/Grafana
docker-compose up -d

# Creacion del cronjob y ejecucion (ejecucion cada 2 minutos)
(crontab -l 2>/dev/null; echo "*/2 * * * * /home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/create_containers.sh") | crontab -

# Carga el modulo de kernel
sudo insmod /home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Modulo_Kernel/modulo.ko