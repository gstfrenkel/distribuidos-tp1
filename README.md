# Steam Analysis

Instrucciones para correr el trabajo:

- Correr docker
- Descargar los [datasets](https://drive.google.com/drive/u/1/folders/1Y2euZUeggfJ9A4Ob5gyj8Cl-p9n_LlsX) y ubicarlos en una carpeta `data` en la raíz del proyecto
- Cada comando a ejecutar está en el Makefile. Por ejemplo:
    - make docker-compose-up
    - make docker-compose-up-client
    - make docker-compose-up-hc
    - make docker-compose-logs
    - make docker-compose-logs-client
    - make docker-compose-logs-hc
    - make docker-compose-down
    - make docker-compose-down-client
    - make docker-compose-down-hc
    - make docker-compose-down-all

Los healthcheckers se deben levantar después de que levante el resto de los servicios del archivo docker-compose.yaml.

En caso de requerir escalar algun nodo, se debe configurar en el docker-compose.yaml. Para correr con más de un cliente, ejecutar:
    
```bash
./client-generator.sh <número de clientes>
```
Esto generará un docker-compose-client.yaml con el número de clientes especificado.

Para abrir el manager de RabbitMQ: http://localhost:15672 con usuario y contraseña `guest`
