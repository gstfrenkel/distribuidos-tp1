import sys


def generate_docker_compose(
    gateway_count,
    reviews_count,
    review_text_count,
    action_count,
    indie_count,
    platform_count,
    joiner_counter_count,
    joiner_top_count,
    joiner_percentile_count,
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
  """

    for i in range(gateway_count):
        docker_compose += f"""
  gateway-{i+1}:
    container_name: gateway-{i+1}
    image: gateway:latest
    environment:
      - worker-id={i}
      - worker-uuid=gateway-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/gateway.toml:/config.toml
      - ./volumes/gateway-{i+1}.csv:/recovery.csv
      - ./volumes/id-generator-{i+1}.csv:/pkg/utils/id/id-generator-{i+1}.csv
"""
    for i in range(reviews_count):
        docker_compose += f"""
  reviews-filter-{i + 1}:
    container_name: reviews-filter-{i + 1}
    image: reviews-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=reviews-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/review.json:/config.json
      - ./volumes/reviews-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(review_text_count):
        docker_compose += f"""
  review-text-filter-{i + 1}:
    container_name: review-text-filter-{i + 1}
    image: review-text-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=review-text-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/text.json:/config.json
      - ./volumes/review-text-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(action_count):
        docker_compose += f"""
  action-filter-{i + 1}:
    container_name: action-filter-{i + 1}
    image: action-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=action-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/action.json:/config.json
      - ./volumes/action-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(indie_count):
        docker_compose += f"""
  indie-filter-{i + 1}:
    container_name: indie-filter-{i + 1}
    image: indie-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=indie-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/indie.json:/config.json
      - ./volumes/indie-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(platform_count):
        docker_compose += f"""
  platform-filter-{i + 1}:
    container_name: platform-filter-{i + 1}
    image: platform-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=platform-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - gateway-1
    networks:
      - tp1_net
    volumes:
      - ./configs/platform.json:/config.json
      - ./volumes/platform-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(joiner_counter_count):
        docker_compose += f"""
  counter-joiner-{i + 1}:
    container_name: counter-joiner-{i + 1}
    image: counter-joiner:latest
    environment:
      - worker-id={i}
      - worker-uuid=counter-joiner-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_counter.json:/config.json
      - ./volumes/counter-joiner-{i+1}.csv:/recovery.csv
"""

    for i in range(joiner_top_count):
        docker_compose += f"""
  top-joiner-{i + 1}:
    container_name: top-joiner-{i + 1}
    image: top-joiner:latest
    environment:
      - worker-id={i}
      - worker-uuid=top-joiner-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_top.json:/config.json
      - ./volumes/top-joiner-{i+1}.csv:/recovery.csv
"""

    for i in range(joiner_percentile_count):
        docker_compose += f"""
  percentile-joiner-{i + 1}:
    container_name: percentile-joiner-{i + 1}
    image: percentile-joiner:latest
    environment:
      - worker-id={i}
      - worker-uuid=percentile-joiner-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/joiner_percentile.json:/config.json
      - ./volumes/percentile-joiner-{i+1}.csv:/recovery.csv
"""

    docker_compose += """
  percentile-aggregator:
    container_name: percentile-aggregator
    image: percentile:latest
    environment:
      - worker-id=0
      - worker-uuid=percentile-aggregator
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/percentile.json:/config.json
      - ./volumes/percentile-aggregator.csv:/recovery.csv
"""

    for i in range(topn_count):
        docker_compose += f"""
  topn-filter-{i + 1}:
    container_name: topn-filter-{i + 1}
    image: topn:latest
    environment:
      - worker-id={i}
      - worker-uuid=topn-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn.json:/config.json
      - ./volumes/topn-filter-{i+1}.csv:/recovery.csv
"""

    docker_compose += """
  topn-aggregator:
    container_name: topn-aggregator
    image: topn:latest
    environment:
      - worker-id=0
      - worker-uuid=topn-aggregator
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_agg.json:/config.json
      - ./volumes/topn-aggregator.csv:/recovery.csv
"""

    for i in range(platform_counter_count):
        docker_compose += f"""
  platform-counter-{i + 1}:
    container_name: platform-counter-{i + 1}
    image: platform-counter:latest
    environment:
      - worker-id={i}
      - worker-uuid=platform-counter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/platform_counter.json:/config.json
      - ./volumes/platform-counter-{i+1}.csv:/recovery.csv
"""

        docker_compose += """
  platform-aggregator:
    container_name: platform-aggregator
    image: platform-counter:latest
    environment:
      - worker-id=0
      - worker-uuid=platform-aggregator
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/platform_counter_agg.json:/config.json
      - ./volumes/platform-aggregator.csv:/recovery.csv
"""

    for i in range(topn_playtime_count):
        docker_compose += f"""
  topn-playtime-filter-{i + 1}:
    container_name: topn-playtime-filter-{i + 1}
    image: topn-playtime-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=topn-playtime-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_playtime.json:/config.json
      - ./volumes/topn-playtime-filter-{i+1}.csv:/recovery.csv
"""

    for i in range(release_date_count):
        docker_compose += f"""
  release-date-filter-{i + 1}:
    container_name: release-date-filter-{i + 1}
    image: release-date-filter:latest
    environment:
      - worker-id={i}
      - worker-uuid=release-date-filter-{i+1}
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/release_date.json:/config.json
      - ./volumes/release-date-filter-{i+1}.csv:/recovery.csv
"""

        docker_compose += """
  topn-playtime-aggregator:
    container_name: topn-playtime-aggregator
    image: topn-playtime-filter:latest
    environment:
      - worker-id=0
      - worker-uuid=topn-playtime-aggregator
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/topn_playtime_agg.json:/config.json
      - ./volumes/topn-playtime-aggregator.csv:/recovery.csv
"""

    docker_compose += """
  review-counter-aggregator:
    container_name: review-counter-aggregator
    image: counter:latest
    environment:
      - worker-id=0
      - worker-uuid=review-counter-aggregator
    depends_on:
      rabbitmq:
        condition: service_healthy
    links:
      - rabbitmq
    networks:
      - tp1_net
    volumes:
      - ./configs/counter_agg.json:/config.json
      - ./volumes/review-counter-aggregator.csv:/recovery.csv

networks:
  tp1_net:
    external: false"""

    return docker_compose


if __name__ == "__main__":

    if len(sys.argv) != 14:
        print(
            "Usage: python3 generate_docker_compose.py <gateway_count> <reviews_count> <review_text_count> <action_count> "
            "<indie_count> <platform_count> <joiner_counter_count> <joiner_top_count> <joiner_percentile_count> "
            "<topn_count> <topn_playtime_count> <release_date_count>"
        )
        sys.exit(1)

    gateway_count = int(sys.argv[1])
    reviews_count = int(sys.argv[2])
    review_text_count = int(sys.argv[3])
    action_count = int(sys.argv[4])
    indie_count = int(sys.argv[5])
    platform_count = int(sys.argv[6])
    joiner_counter_count = int(sys.argv[7])
    joiner_top_count = int(sys.argv[8])
    joiner_percentile_count = int(sys.argv[9])
    topn_count = int(sys.argv[10])
    platform_counter_count = int(sys.argv[11])
    topn_playtime_count = int(sys.argv[12])
    release_date_count = int(sys.argv[13])

    docker_compose = generate_docker_compose(
        gateway_count,
        reviews_count,
        review_text_count,
        action_count,
        indie_count,
        platform_count,
        joiner_counter_count,
        joiner_top_count,
        joiner_percentile_count,
        topn_count,
        topn_playtime_count,
        release_date_count,
    )

    with open("../docker-compose.yaml", "w") as f:
        f.write(docker_compose)

    print("Docker Compose configuration saved to docker-compose.yaml")
