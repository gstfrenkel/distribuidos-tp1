# Explicación de los scripts disponibles

1. **`client-generator.sh`**: Genera los docker compose para N clientes
2. **`generate-docker-compose.sh`**: Genera el docker compose del sistema
3. **`hc-generator.sh`**: Genera los docker compose para N health checks, incluyendo los .env necesarios para estos. CORRER DESDE EL DIR SCRIPTS!
4. **`random-kill.sh`**: Mata un contenedor aleatorio de la red cada 35 segundos
5. **`run-comparison.sh`**: Compara dos archivos de resultados 
6. **`recovery-files-creator.py`**: Crea los archivos de recuperación necesarios para el sistema (esto es porque docker compose los necesita creados para montarlos como volumen).