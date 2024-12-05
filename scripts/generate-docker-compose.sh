#!/bin/bash

read -p "Enter the number of gateway instances: " gateway_count
read -p "Enter the number of reviews filter instances: " reviews_count
read -p "Enter the number of review text filter instances: " review_text_count
read -p "Enter the number of action filter instances: " action_count
read -p "Enter the number of indie filter instances: " indie_count
read -p "Enter the number of platform filter instances: " platform_count
read -p "Enter the number of joiner counter filter instances: " joiner_counter_count
read -p "Enter the number of joiner top filter instances: " joiner_top_count
read -p "Enter the number of joiner percentile filter instances: " joiner_percentile_count
read -p "Enter the number of topn filter instances: " topn_count
read -p "Enter the number of platform counter instances: " platform_counter_count
read -p "Enter the number of topn playtime filter instances: " topn_playtime_count
read -p "Enter the number of release date filter instances: " release_date_count

python3 generate_docker_compose.py "$gateway_count" "$reviews_count" "$review_text_count" "$action_count" "$indie_count" "$platform_count" "$joiner_counter_count" "$joiner_top_count" "$joiner_percentile_count" "$topn_count" "$platform_counter_count" "$topn_playtime_count" "$release_date_count"
