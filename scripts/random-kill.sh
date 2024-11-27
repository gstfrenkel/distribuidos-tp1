#!/bin/bash

containers_exclude=("client" "rabbitmq")
list_containers() {
    docker network inspect distribuidos-tp1_tp1_net -f '{{range .Containers}}{{.Name}}{{"\n"}}{{end}}' | grep -v -E "$(IFS=\|; echo "${containers_exclude[*]}")"
}


main() {
    KILL_INTERVAL=${1:-35}

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