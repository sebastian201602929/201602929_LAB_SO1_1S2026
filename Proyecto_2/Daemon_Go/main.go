package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ContainerInfo struct {
	pid    string
	nombre string
	cmd    string
	vsz    string
	rss    string
	rampc  string
	cpupc  string
	clasif string
}

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

		if campos[2] == "roldyoran/go-client" || campos[2] == "alpine sh -c \"while true; do echo '2^20' | bc > /dev/null; sleep 2; done\"" {
			clasificacion = "Alto consumo"
		} else if campos[2] == "alpine sleep 240" {
			clasificacion = "Bajo consumo"
		} else if campos[2] == "Grafana" {
			clasificacion = "Grafana"
		} else if campos[2] == "Valkey" {
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
			clasif: clasificacion,
		}

		containers = append(containers, c)

		// Incrementa numero de linea
		//line_num++ // Es irrelevanta cuando ya llego a la 5 linea, ya que las demas siempre tendran info de contenedor
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	// Ordenar contenedores

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
	} else if image == "grafana/grafana" {
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

// Función para ordenar contenedores por CPU de menor a mayor
func sortContainersByCPU(containers []ContainerInfo) {
	sort.Slice(containers, func(i, j int) bool {
		cpuI, errI := strconv.ParseFloat(containers[i].cpupc, 64)
		cpuJ, errJ := strconv.ParseFloat(containers[j].cpupc, 64)

		if errI != nil || errJ != nil {
			log.Printf("Error al parsear CPU para ordenamiento: %v, %v", errI, errJ)
			return false
		}

		return cpuI < cpuJ
	})
}

func main() {
	// Abre archivo de log
	f, err := os.OpenFile("daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)

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

	// Loop principal
	for i := 0; i < 2; i++ {
		// Lee archivo /proc/continfo_pr2_so1_201602929 y deserializa data
		containers := readContainerInfo()

		if containers == nil {
			log.Printf("No se pudo leer el archivo del modulo, saltando iteracion...")

			time.Sleep(20 * time.Second)

			continue
		}

		// Ordena los contenedores por

		// Analiza data y decide si detener contenedores
		// Generacion de logs en Valkey, para que consuma Grafana

		log.Printf("Daemon activo a las %s", time.Now().Format(time.RFC1123))

		time.Sleep(20 * time.Second)
	}

	// Finaliza servicio
	// Limpieza de agregados
	cmd = exec.Command("bash", "/home/sebastian/Desktop/Sopes_1/201602929_LAB_SO1_1S2026/Proyecto_2/Daemon_Go/down_script.sh")
	down_err := cmd.Run()

	if down_err != nil {
		log.Fatalf("Error al bajar los prerrequisitos")
		panic(down_err)
	}
}
