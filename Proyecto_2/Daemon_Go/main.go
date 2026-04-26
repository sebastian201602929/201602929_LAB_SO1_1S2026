package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type ContainerInfo struct {
	pid    string
	nombre string
	cmd    string
	vsz    string
	rss    string
	rampc  string
	cpupc  string
	cid    string
	clasif string
}

var RAM_total string
var RAM_uso string
var contenedores_detenidos int

func readContainerInfo() []ContainerInfo {
	// Abre archivo del modulo
	f, err := os.Open("/proc/continfo_pr2_so1_201602929")

	if err != nil {
		return nil
	}

	defer f.Close()

	var containers []ContainerInfo

	// Crea scanner del archivo
	scanner := bufio.NewScanner(f)

	// Contador de lineas para saltar las primeras
	line_num := 0

	// Ciclo para leer las lineas
	for scanner.Scan() {
		// Salta las primeras 5 lineas (3 Info del sistema, un salto, y encabezados de la tabla)
		if line_num == 0 {
			line_num++

			// Leer Total de RAM
			total_ram_line := scanner.Text()
			total_ram_parts := strings.Split(total_ram_line, ":")

			total_ram := strings.TrimSpace(total_ram_parts[1])
			RAM_total = strings.Split(total_ram, " ")[0]

			continue
		} else if line_num == 2 {
			line_num++

			// Leer RAM en uso
			used_ram_line := scanner.Text()
			used_ram_parts := strings.Split(used_ram_line, ":")

			// Formato de la linea: "Memoria RAM en uso: 5697 MB"
			used_ram := strings.TrimSpace(used_ram_parts[1])
			RAM_uso = strings.Split(used_ram, " ")[0]

			continue
		}

		if line_num < 5 {
			line_num++
			continue
		}

		// Obtiene linea
		linea := scanner.Text()

		// Separa la linea por tabs
		campos := strings.Split(linea, "\t")

		if len(campos) != 7 {
			continue // Ignorar lineas malas
		}

		cmd_campos := strings.Split(campos[2], " ")

		cid := cmd_campos[4] // El id del contenedor esta en el campo 4

		clasificacion := getDockerImage(cid)

		if clasificacion == "roldyoran/go-client" || clasificacion == "alpine sh -c while true; do echo '2^20' | bc > /dev/null; sleep 2; done" {
			clasificacion = "Alto consumo"
		} else if clasificacion == "alpine sleep 240" {
			clasificacion = "Bajo consumo"
		} else if clasificacion == "Grafana" {
			clasificacion = "Grafana"
		} else if clasificacion == "Valkey" {
			clasificacion = "Valkey"
		} else {
			clasificacion = "Desconocido"
		}

		c := ContainerInfo{
			pid:    campos[0],
			nombre: campos[1],
			cmd:    campos[2],
			vsz:    campos[3],
			rss:    campos[4],
			rampc:  campos[5],
			cpupc:  campos[6],
			cid:    cid,
			clasif: clasificacion,
		}

		containers = append(containers, c)

		// Incrementa numero de linea
		//line_num++ // Es irrelevanta cuando ya llego a la 5 linea, ya que las demas siempre tendran info de contenedor
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	// Ordena contenedores
	sort.Slice(containers, func(i, j int) bool {
		// Ordena por uso de RAM
		ramI, errI := strconv.ParseFloat(containers[i].rampc, 64)
		ramJ, errJ := strconv.ParseFloat(containers[j].rampc, 64)

		if errI != nil || errJ != nil {
			log.Printf("Error al parsear RAM para ordenamiento: %v, %v", errI, errJ)
			return false
		}

		if ramI != ramJ {
			return ramI < ramJ
		}

		// Si la RAM es igual, ordena por VSZ
		vszI, errI := strconv.ParseFloat(containers[i].vsz, 64)
		vszJ, errJ := strconv.ParseFloat(containers[j].vsz, 64)

		if errI != nil || errJ != nil {
			log.Printf("Error al parsear VSZ para ordenamiento: %v, %v", errI, errJ)
			return false
		}

		if vszI != vszJ {
			return vszI < vszJ
		}

		// Si la RAM y VSZ son iguales, ordena por RSS
		rssI, errI := strconv.ParseFloat(containers[i].rss, 64)
		rssJ, errJ := strconv.ParseFloat(containers[j].rss, 64)

		if errI != nil || errJ != nil {
			log.Printf("Error al parsear RSS para ordenamiento: %v, %v", errI, errJ)
			return false
		}

		if rssI != rssJ {
			return rssI < rssJ
		}

		// Si la RAM, VSZ y RSS son iguales, ordena por uso de CPU
		cpuI, errI := strconv.ParseFloat(containers[i].cpupc, 64)
		cpuJ, errJ := strconv.ParseFloat(containers[j].cpupc, 64)

		if errI != nil || errJ != nil {
			log.Printf("Error al parsear CPU para ordenamiento: %v, %v", errI, errJ)
			return false
		}

		return cpuI < cpuJ
	})

	return containers
}

// Función para obtener la imagen Docker del contenedor
func getDockerImage(cid string) string {
	// Obtiene la imagen del contenedor
	cmd_image := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", cid)

	var out_image bytes.Buffer
	cmd_image.Stdout = &out_image
	err_image := cmd_image.Run()

	if err_image != nil {
		return ""
	}

	image := strings.TrimSpace(out_image.String())

	if image == "roldyoran/go-client" {
		return "roldyoran/go-client"
	} else if image == "grafana-grafana" {
		return "Grafana"
	} else if image == "valkey/valkey:latest" {
		return "Valkey"
	}

	// Obtiene el comando con el que fue creado el contenedor
	cmd_cmd := exec.Command("docker", "inspect", "--format", "{{json .Config.Cmd}}", cid)

	var out_cmd bytes.Buffer
	cmd_cmd.Stdout = &out_cmd
	err_cmd := cmd_cmd.Run()

	if err_cmd != nil {
		return ""
	}

	// Este comando devuelve en un json todo lo que se envio al crear el contenedor
	// Se convierte a un string
	var cmdSlice []string
	err_cmd = json.Unmarshal([]byte(strings.TrimSpace(out_cmd.String())), &cmdSlice)
	if err_cmd != nil {
		return ""
	}

	cmd := strings.Join(cmdSlice, " ")

	return image + " " + cmd
}

func main() {
	// Abre archivo de log
	f, err := os.OpenFile("daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Print("Iniciando servicio...")

	// Preparación del servicio
	// Crea contenedor de Grafana con Valkey, crea cronjob y carga modulo de kernel
	cmd := exec.Command("bash", "/home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/up_script.sh")
	up_err := cmd.Run()

	if up_err != nil {
		log.Fatalf("Error al levantar los prerrequisitos del servicio... Abortando")
		log.Fatal(up_err)

		// Ejecucion de down_script
		cmd = exec.Command("bash", "/home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/down_script.sh")
		down_err := cmd.Run()

		if down_err != nil {
			log.Fatalf("Alm, fallo el script que aborta también! :'(")
			panic(down_err)
		}

		return
	}

	log.Print("up_script.sh ejecutado correctamente! :3")

	// Canal para recibir la señal de terminación del servicio
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Canal para indicar que el programa debe terminar
	done := make(chan bool, 1)

	go func() {
		sig := <-quit
		log.Printf("Señal de salida, apagando servicio...", sig)

		// Ejecuta down_script.sh
		cmd = exec.Command("bash", "/home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/down_script.sh")
		down_err := cmd.Run()

		if down_err != nil {
			log.Fatalf("Error al bajar las dependencias del servicio")
			panic(down_err)
		}

		done <- true
	}()

	// Crea cliente de Valkey
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Limpia los datos anteriores
	rdb.FlushDB(ctx)

	// Loop principal
	for {
		// Lee archivo /proc/continfo_pr2_so1_201602929 y deserializa data
		containers := readContainerInfo()

		if containers == nil {
			log.Printf("No se pudo leer el archivo del modulo, saltando iteracion...")

			time.Sleep(20 * time.Second)

			continue
		}

		alto_consumo := 0
		bajo_consumo := 0

		// Recorre los contenedores y decide cuales detener y/o eliminar
		for _, c := range containers {
			// Los contenedores de Grafana y Valkey no se tocan
			if c.clasif == "Grafana" || c.clasif == "Valkey" {
				continue
			} else if c.clasif == "Alto consumo" {
				if alto_consumo < 2 {
					alto_consumo++
					continue
				}
			} else if c.clasif == "Bajo consumo" {
				if bajo_consumo < 3 {
					bajo_consumo++
					continue
				}
			}

			// Generar log o algo no ? xd
			//log.Printf("Contenedor PID: %s, Nombre: %s, Comando: %s, VSZ: %s, RSS: %s, RAM%%: %s, CPU%%: %s, CID: %s, Clasificacion: %s", c.pid, c.nombre, c.cmd, c.vsz, c.rss, c.rampc, c.cpupc, c.cid, c.clasif)

			// Detiene el contenedor
			cmd := exec.Command("docker", "stop", c.cid)
			err := cmd.Run()

			if err != nil {
				log.Printf("Error al detener contenedor %s: %v", c.cid, err)
			}

			contenedores_detenidos++

			// Conversion de valores de RAM% y CPU% a float
			rampc, _ := strconv.ParseFloat(c.rampc, 64)
			cpupc, _ := strconv.ParseFloat(c.cpupc, 64)

			// Inserta los valores del contenedor en Valkey para los tops
			rdb.ZAdd(ctx, "top_ram", redis.Z{Score: rampc, Member: c.cid})
			rdb.ZAdd(ctx, "top_cpu", redis.Z{Score: cpupc, Member: c.cid})
		}

		// Calcula RAM libre
		i_RAM_total, _ := strconv.Atoi(RAM_total)
		i_RAM_uso, _ := strconv.Atoi(RAM_uso)
		RAM_libre := strconv.Itoa((i_RAM_total - i_RAM_uso))

		// Generacion de logs en Valkey, para que consuma Grafana
		rdb.Set(ctx, "ram_total", RAM_total, 0)
		rdb.Set(ctx, "ram_usada", RAM_uso, 0)
		rdb.Set(ctx, "ram_libre", RAM_libre, 0)
		rdb.Set(ctx, "contenedores_detenidos", strconv.Itoa(contenedores_detenidos), 0)

		log.Printf("%s: Servicio iteró correctamente.", time.Now().Format(time.RFC1123))

		time.Sleep(20 * time.Second)
	}

	<-done
}
