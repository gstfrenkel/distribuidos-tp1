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
    KILL_INTERVAL=10
    KILL_COUNT=5

    while true; do
        containers=($(list_containers))
        n_hc=0
        if [ ${#containers[@]} -gt 0 ]; then
            for ((i=0; i<KILL_COUNT && i<${#containers[@]}; i++)); do
                            container_to_kill=${containers[$RANDOM % ${#containers[@]}]}

                            if [[ "$container_to_kill" =~ healthchecker_.* ]]; then
                                n_hc=$n_hc + 1
                                if [ $n_hc -gt $KILL_COUNT-1 ]; then
                                    n_hc=0
                                    continue
                                fi
                            fi
                            docker kill "$container_to_kill" &>/dev/null
                            echo "Container $container_to_kill has been killed."
                        done
        else
            echo "No containers to kill."
        fi

        sleep "$KILL_INTERVAL"
    done
}

main "$@"