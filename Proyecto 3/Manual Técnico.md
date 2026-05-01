# Manual Técnico

## Arquitectura general del Proyecto

El proyecto 3 consta de un flujo en Kubernetes para archivos publicados en formato JSON. Los datos son generados aleatoriamente y en masivo por Locust corriendo en la maquina local. A partir de este punto, todo corre en un cluster de Kubernetes. Con ayuda de un Gateway API, se publican en una API Rust en Kubernetes. Esta API Rust se comunica y publica los datos en una API Go, que a su vez funciona como gRPC Client. Este cliente, publica los datos en el gRPC Server, el cual a su vez funciona como publisher de RabbitMQ. Por otro lado, se tiene un pod de Consumer de RabbitMQ en Go, el cual publica los datos en un servicio de Valkey que corre en una VM creada con Kubevirt en el cluster. Esta VM con Valkey es consuida por otra VM creada con Kubevirt que corre Grafana. Finalmente, se tiene un NodePort, que nos permite conectarnos a Grafana para poder visualizar los datos que hay fluido por el proyecto.

Todos estos deploys, se realizan obteniendo las imagenes desde un repositorio privado de artefactos OCI, una VM externa al cluster de Kubernetes, que corre Zot.

## Flujo completo de datos

Locust (generación de los datos) -> API Rust -> gRPC Client -> gRPC Server -> RabbitMQ -> Consumer -> Valkey -> Grafana

## Configuración de Gateway API

Se ha configurado el Gateway API desde el yaml de api-rust, para minimizar los errores de despliegue por dependencias. Se adjunta la sección de código correspondiente.

```
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-rust-route
  namespace: mumnk8s
spec:
  parentRefs:
    - name: api-gateway
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      backendRefs:
        - name: api-rust-service
          port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: api-gateway
  namespace: mumnk8s
spec:
  gatewayClassName: gke-l7-global-external-managed
  listeners:
    - name: http
      protocol: HTTP
      port: 80
```

## Comunicación REST y gRPC

La API desarrollada en Rust funciona totalmente como una API REST. Esta se comunica con el cliente gRPC, el cual corre tambien una API REST para recibir los datos. En esta comunicación, los datos se encuentran en formato JSON.

Luego, el gRPC Client traslada los datos hacia el gRPC Server en formato binario. Esto ayuda a que la comunicación sea rápida.

## Uso de RabbitMQ

RabbitMQ se despliega con las opciones por defecto desde la imagen oficial de RabbitMQ. El gRPC Server se conecta y crea un canal. Este crea la cola ```war_reports``` y publica todos los datos que obtiene.

Por otro lado, el consumer obtiene los datos de esta misma cola y los publica en Valkey.

## Despliegue de Valkey y Grafana sobre Kubevirt

Debido a que Kubevirt nos permite crear máquinas virtuales en Kubernetes, se han creado 2 VM para cumplir con esta necesidad del proyecto.

En la primera VM se corre Fedora y se ha instalado Valkey como un servicio local.

En la segunda VM se corre Alpine y se ha instalado Grafana en un disco adicional acoplado con mayor capacidad. Luego, se levanta Grafana como un servicio.

## Configuración de HPA

Se ha configurado el deploy de API Rust para ser escalable horizontalmente cuando se suera un porcentaje de utilización de cpu del 30%, permitiéndo crear hasta 3 réplicas. Esto se ha realizado en el mismo yaml de api-rust, para evitar errores de depndencias. Se adjunta la sección de código.

```
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-rust-hpa
  namespace: mumnk8s
spec:
  minReplicas: 1
  maxReplicas: 3
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-rust-deploy
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 30
```

## Publicación/Consumo de imágenes desde Zot

Las imágenes de las apps desplegadas en Kubernetes se han publicado en un repositorio privado de artefactos OCI. Se ha utilizado Zot corriendo en una máquina virtual externa al cluster de Kubernetes. En esta máquina virtual se ha instalado el servicio de Ngrok para crear una URL que no cambia y funciona con protocolo HTTPS, permitiendo ser accedida por GKE y aliviando la preocupación de que la IP cambie cada vez que necesitemos cargar nuestras imágenes.

---

# Pruebas realizadas

Se ha probado el flujo 2 veces.

La primera prueba se realizó en local. Se ha probado el flujo desde Locust hasta gRPC Server. No tengo una captura de esa prueba.

La segunda prueba ha sido al finalizar el proyecto y desplegarlo en GKE. En esta ocasión, para no estresar el sistema, sólo se han enviado 2 datos, para poder probar también el dashboard de grafana.

![Envío de datos en JSON hacía el Gateway API](/Proyecto%203/img/post_consola.png)

![Dashboard de Grafana con los datos recibidos](/Proyecto%203/img/grafana.png)

---

# Guía de Instalación

## Locust

Entrar al venv. ```source venv/bin/activate```.

Se corre: ```locust -f locust.py --host=http://34.95.98.114/grpc-201602929```. Esto puede cambiar si se apaga el cluster por el Gateway API.

Dentro del venv se necesita tener instalado locust, o se instala con ```pip install locust```

Para ver la pagina de locust ```http://localhost:8089/```.

## API Rust

Para ejecutar la API ```cargo run```.
Pero como ahora es un contenedor, solo ```docker run -p 8080:8080 api-rust:v1```
Tag: ```docker tag api-rust:v1 puppy-imaging-props.ngrok-free.dev/api-rust:v1```
Push: ```docker push puppy-imaging-props.ngrok-free.dev/api-rust:v1```

## Deployment 1 de Go - gRPC Client

Cliente Go de gRPC. Se buildea el Dockerfile desde la carpeta raiz del Proyecto 3.

```
docker build -t grpc-client:v1 -f API-gRPC-Client/Dockerfile .
docker tag grpc-client:v1 puppy-imaging-props.ngrok-free.dev/grpc-client:v1
docker push puppy-imaging-props.ngrok-free.dev/grpc-client:v1
```

## Deployment 2 de Go - gRPC Server

Server gRPC. Se buildea el Dockerfile desde la carpeta raiz del Proyecto 3.

```
docker build -t grpc-server:v1 -f gRPC-Server/Dockerfile .
docker tag grpc-server:v1 puppy-imaging-props.ngrok-free.dev/grpc-server:v1
docker push puppy-imaging-props.ngrok-free.dev/grpc-server:v1
```

## Deployment 3 de Go - RabbitMQ Consumer

Solo se crea la imagen: ```docker build -t consumer:v1 .```
Tag: ```docker tag consumer:v1 puppy-imaging-props.ngrok-free.dev/consumer:v1```
Push: ```docker push puppy-imaging-props.ngrok-free.dev/consumer:v1```

## Kubevirt

Practicamente, ya estan aplicados los yamls, pero por si acaso lo necesito, primero aplicar el ```kubectl apply -f k8s/kubevirt/emul.yaml``` paara crear el operador.

### Grafana VM

Aplicar ```kubectl apply -f k8s/kubevirt/grafana-vm.yaml``` para crear la vm de Grafana. Basado en el ejemplo del aux.

Para acceder a la vm, primero validamos que este lista con ```kubectl get vm -n mumnk8s``` y ```kubectl get vmi -n mumnk8s```
Si estan en Running, ya me puedo conectar.

Conectarse con ```virtctl console grafana-vm -n mumnk8s```. El usuario es ```root```.

Levantar la red, siempre se hace ```ip link set eth0 up && udhcpc -i eth0```

Formatear disco extra y montarlo
```
mkdosfs -F 32 /dev/vdb
mkdir /mnt/data
mount /dev/vdb /mnt/data
```

Instalar Grafana
```
wget https://dl.grafana.com/oss/release/grafana-11.5.0.linux-amd64.tar.gz -O /mnt/data/grafana.tar.gz
tar -zxvf /mnt/data/grafana.tar.gz -C /mnt/data/
mkdir -p /mnt/data/grafana-data /mnt/data/grafana-logs /mnt/data/grafana-plugins
```

Iniciar servicio de Grafana
```
/mnt/data/grafana-v11.5.0/bin/grafana-server \
  --homepath /mnt/data/grafana-v11.5.0 \
  cfg:default.paths.data=/mnt/data/grafana-data \
  cfg:default.paths.logs=/mnt/data/grafana-logs \
  cfg:default.paths.plugins=/mnt/data/grafana-plugins &
```

### Valkey VM

Aplicar ```kubectl apply -f k8s/kubevirt/valkey-vm.yaml``` para crear la vm de Valkey.

Para acceder a la vm, primero validamos que este lista con ```kubectl get vm -n mumnk8s``` y ```kubectl get vmi -n mumnk8s```
Si estan en Running, ya me puedo conectar.

Conectarse con ```virtctl console valkey-vm -n mumnk8s```. El usuario y password son ```fedora```.

Instalar Valkey
```
sudo dnf update -y
sudo dnf install -y gcc make wget tar
wget https://github.com/valkey-io/valkey/archive/refs/tags/7.2.5.tar.gz -O valkey.tar.gz
tar -xzf valkey.tar.gz
cd valkey-7.2.5
make
```

Iniciar servicio Valkey

```
./src/valkey-server --bind 0.0.0.0 --protected-mode no --requirepass 201602929 --daemonize yes
sudo sysctl vm.overcommit_memory=1
```

## VM con Zot

Al encenderla, correr el contenedor de Zot: ```docker start zot```.
Correr Ngrok para tener la url que no cambia y que sea segura: ```ngrok http 5000```.

## Crear cluster de GKE

```
gcloud container clusters create mumnk8s-cluster \
  --zone us-central1-a \
  --num-nodes 4 \
  --machine-type e2-standard-4 \
  --gateway-api standard \
  --disk-size=100
```

Configurar acceso

```
gcloud container clusters get-credentials mumnk8s-cluster --zone us-central1-a
```

Aplicar yamls

```
kubectl apply -f k8s/
```

Para que no consuma creditos el cluster y no tener que configurarlo de nuevo

```
gcloud container clusters resize mumnk8s-cluster \
  --node-pool default-pool \
  --num-nodes 0 \
  --zone us-central1-a
```

Se suele quedar trabado, con este comando se fuerza a detener las vm (los nodos) manualmente
```
gcloud compute instances delete gke-mumnk8s-cluster-default-pool-99c6f7ec-48dc --zone us-central1-a --quiet
```

---

# Conclusiones

Fue un proyecto interesante. Kubernetes es una herramienta que nos permite desplegar microservicios como si trabajaramos con contenedores localmente, pero permitiendo la escalabilidad de los sistemas.

Por otra parte, no me agrada Kubevirt. No sé si no lo he entendido por completo, pero me parece una solución que limita en gran parte los recursos, además que Kubernetes ya despliega contenedores.