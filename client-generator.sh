#!/bin/bash

if [ -z "$1" ]; then
  echo "Uso: $0 <numero_de_clientes>"
  exit 1
fi

NUM_CLIENTS=$1

OUTPUT_FILE="docker-compose-client.yaml"

cat > $OUTPUT_FILE <<EOL
version: '3'
services:
EOL

for ((i=1; i<=NUM_CLIENTS; i++)); do
  cat >> $OUTPUT_FILE <<EOL
  client-$i:
    container_name: client-$i
    image: client:latest
    volumes:
      - ./data:/app/data
      - ./configs/client.toml:/app/config.toml
    networks:
      - tp1_net
EOL
done

cat >> $OUTPUT_FILE <<EOL
networks:
  tp1_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOL

echo "Archivo docker-compose-client.yaml generado con $NUM_CLIENTS clientes."
