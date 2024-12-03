#!/bin/bash

if [ -z "$1" ]; then
    echo "Error! Usage: $0 <network>"
    exit 1
fi

NETWORK=$1

containers_exclude=("client" "rabbitmq")
list_containers() {
    docker network inspect $NETWORK -f '{{range .Containers}}{{.Name}}{{"\n"}}{{end}}' | grep -v -E "$(IFS=\|; echo "${containers_exclude[*]}")"
}


main() {
    KILL_INTERVAL=35

    while true; do
        containers=($(list_containers))

        if [ ${#containers[@]} -gt 0 ]; then
            container_to_kill=${containers[$RANDOM % ${#containers[@]}]}
            docker kill "$container_to_kill"
            echo "Container $container_to_kill has been killed."
        else
            echo "No containers to kill."
        fi

        sleep "$KILL_INTERVAL"
    done
}

main "$@"