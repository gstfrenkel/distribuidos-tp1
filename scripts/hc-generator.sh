#!/bin/bash

if [ -z "$1" -o -z "$2" ]; then
  echo "Error! Usage: $0 <number_of_healthcheckers> <net>"
  exit 1
fi

NUM_HEALTHCHECKERS=$1
NETWORK=$2
OUTPUT_FILE="../docker-compose-hc.yaml"

# Función para listar contenedores en la red excluyendo ciertos nombres
containers_exclude=("client" "rabbitmq")
list_containers() {
    docker network inspect $NETWORK -f '{{range .Containers}}{{.Name}}{{"\n"}}{{end}}' | grep -v -E "$(IFS=\|; echo "${containers_exclude[*]}")"
}

# Obtener la lista de contenedores
readarray -t CONTAINERS < <(list_containers)
TOTAL_CONTAINERS=${#CONTAINERS[@]}

if [ $TOTAL_CONTAINERS -eq 0 ]; then
  echo "No containers found in the network $NETWORK"
  exit 1
fi

CONTAINERS_PER_CHECKER=$((TOTAL_CONTAINERS / NUM_HEALTHCHECKERS))
REMAINING_CONTAINERS=$((TOTAL_CONTAINERS % NUM_HEALTHCHECKERS))

# Generar el archivo docker-compose
cat > $OUTPUT_FILE <<EOL
services:
EOL

for ((i=1; i<=NUM_HEALTHCHECKERS; i++)); do
  cat >> $OUTPUT_FILE <<EOL
  healthchecker-$i:
    container_name: healthchecker-$i
    image: healthchecker:latest
    env_file: ./configs/env/healthchecker-$i.env
    networks:
      - tp1_net
    volumes:
      - ./configs/healthchecker.toml:/app/config.toml
      - /var/run/docker.sock:/var/run/docker.sock
EOL
done

cat >> $OUTPUT_FILE <<EOL
networks:
  tp1_net:
    external: false
EOL
echo "docker-compose-hc.yaml with $NUM_HEALTHCHECKERS healthcheckers generated"

# Generar los archivos .env para cada healthchecker
# Deben estar corriendo los contenedores para que esto funcione
for ((i=1; i<=NUM_HEALTHCHECKERS; i++)); do
  ENV_FILE="../configs/env/healthchecker-$i.env"
  NEXT=$((i % NUM_HEALTHCHECKERS + 1))

  START_INDEX=$(( (i - 1) * CONTAINERS_PER_CHECKER ))
  END_INDEX=$((START_INDEX + CONTAINERS_PER_CHECKER))

  if (( i <= REMAINING_CONTAINERS )); then
    END_INDEX=$((END_INDEX + 1))
  fi

  NODES_LIST="${CONTAINERS[@]:$START_INDEX:$((END_INDEX - START_INDEX))}"

  cat > $ENV_FILE <<EOL
id=$i
next=$NEXT
nodes=$NODES_LIST
EOL
done