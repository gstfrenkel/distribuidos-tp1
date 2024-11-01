import sys


def generate_docker_compose(
    reviews_count,
    review_text_count,
    action_count,
    indie_count,
    platform_count,
    joiner_counter_count,
    joiner_top_count,
    joiner_percentile_count,
    percentile_count,
    topn_count,
    topn_playtime_count,
    release_date_count,
):
    docker_compose = """services:
  rabbitmq:
    container_name: rabbitmq
    image: rabbitmq:management
    ports:
      - "15672:15672"
    networks:
      - tp1_net
    healthcheck:
      test: rabbitmq-diagnostics check_port_connectivity
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 50s
    logging:
      driver: none

  gateway-1:
    container_name: gateway-1
    image: gateway:latest
    environment:
      - worker-id=0
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/gateway.toml:/app/config.toml
"""
    for i in range(reviews_count):
        docker_compose += f"""
  reviews-filter-{i + 1}:
    container_name: reviews-filter-{i + 1}
    image: reviews-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/review.json:/app/config.json
"""

    for i in range(review_text_count):
        docker_compose += f"""
  review-text-filter-{i + 1}:
    container_name: review-text-filter-{i + 1}
    image: review-text-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/text.json:/app/config.json
"""

    for i in range(action_count):
        docker_compose += f"""
  action-filter-{i + 1}:
    container_name: action-filter-{i + 1}
    image: action-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/action.json:/app/config.json
"""

    for i in range(indie_count):
        docker_compose += f"""
  indie-filter-{i + 1}:
    container_name: indie-filter-{i + 1}
    image: indie-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/indie.json:/app/config.json
"""

    for i in range(platform_count):
        docker_compose += f"""
  platform-filter-{i + 1}:
    container_name: platform-filter-{i + 1}
    image: platform-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - gateway-1
    networks:
      - tp1_net
    volumes:
      - ./configs/platform.json:/app/config.json
"""

    for i in range(joiner_counter_count):
        docker_compose += f"""
  joiner-counter-filter-{i + 1}:
    container_name: joiner-counter-filter-{i + 1}
    image: joiner-counter-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_counter.json:/app/config.json
"""

    for i in range(joiner_top_count):
        docker_compose += f"""
  joiner-top-filter-{i + 1}:
    container_name: joiner-top-filter-{i + 1}
    image: joiner-top-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_top.json:/app/config.json
"""

    for i in range(joiner_percentile_count):
        docker_compose += f"""
  joiner-percentile-filter-{i + 1}:
    container_name: joiner-percentile-filter-{i + 1}
    image: joiner-percentile-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_percentile.json:/app/config.json
"""

    for i in range(percentile_count):
        docker_compose += f"""
  percentile-filter-{i + 1}:
    container_name: percentile-filter-{i + 1}
    image: percentile:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/percentile.json:/app/config.json
"""

    for i in range(topn_count):
        docker_compose += f"""
  topn-filter-{i + 1}:
    container_name: topn-filter-{i + 1}
    image: topn:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn.json:/app/config.json
"""

    docker_compose += """
  topn-aggregator-1:
    container_name: topn-aggregator-1
    image: topn:latest
    environment:
      - worker-id=0
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_agg.json:/app/config.json
"""

    for i in range(platform_counter_count):
        docker_compose += f"""
  platform-counter-{i + 1}:
    container_name: platform-counter-{i + 1}
    image: platform-counter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/platform_counter.json:/app/config.json
"""

        docker_compose += """
  platform-aggregator-1:
    container_name: platform-aggregator-1
    image: platform-counter:latest
    environment:
      - worker-id=0
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/platform_counter_agg.json:/app/config.json
"""

    for i in range(topn_playtime_count):
        docker_compose += f"""
  topn-playtime-filter-{i + 1}:
    container_name: topn-playtime-filter-{i + 1}
    image: topn-playtime-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_playtime.json:/app/config.json
"""

    for i in range(release_date_count):
        docker_compose += f"""
  release-date-filter-{i + 1}:
    container_name: release-date-filter-{i + 1}
    image: release-date-filter:latest
    environment:
      - worker-id={i}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/release_date.json:/app/config.json
"""

        docker_compose += """
  topn-playtime-aggregator-1:
    container_name: topn-playtime-aggregator-1
    image: topn-playtime-filter:latest
    environment:
      - worker-id=0
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_playtime_agg.json:/app/config.json
"""

    docker_compose += """
  review-counter-agg:
    container_name: review-counter-agg
    image: counter:latest
    environment:
      - worker-id=0
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/counter_agg.json:/app/config.json

networks:
  tp1_net:
    external: false
"""

    return docker_compose


if __name__ == "__main__":

    if len(sys.argv) != 14:
        print(
            "Usage: python3 generate_docker_compose.py <reviews_count> <review_text_count> <action_count> "
            "<indie_count> <platform_count> <joiner_counter_count> <joiner_top_count> <joiner_percentile_count> "
            "<percentile_count> <topn_count> <topn_playtime_count> <release_date_count>"
        )
        sys.exit(1)

    reviews_count = int(sys.argv[1])
    review_text_count = int(sys.argv[2])
    action_count = int(sys.argv[3])
    indie_count = int(sys.argv[4])
    platform_count = int(sys.argv[5])
    joiner_counter_count = int(sys.argv[6])
    joiner_top_count = int(sys.argv[7])
    joiner_percentile_count = int(sys.argv[8])
    percentile_count = int(sys.argv[9])
    topn_count = int(sys.argv[10])
    platform_counter_count = int(sys.argv[11])
    topn_playtime_count = int(sys.argv[12])
    release_date_count = int(sys.argv[13])

    docker_compose = generate_docker_compose(
        reviews_count,
        review_text_count,
        action_count,
        indie_count,
        platform_count,
        joiner_counter_count,
        joiner_top_count,
        joiner_percentile_count,
        percentile_count,
        topn_count,
        topn_playtime_count,
        release_date_count,
    )

    with open("docker-compose.yaml", "w") as f:
        f.write(docker_compose)

    print("Docker Compose configuration saved to docker-compose.yaml")
