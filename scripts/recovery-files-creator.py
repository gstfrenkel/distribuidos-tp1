import re
import os

with open('../docker-compose.yaml', 'r') as file:
    docker_compose_content = file.read()

os.makedirs('../volumes', exist_ok=True)
volume_pattern = re.compile(r'- \./volumes/[^:]+') # regex to match volumes in the form of - ./volumes/...
volumes = volume_pattern.findall(docker_compose_content)

for volume in volumes:
    host_path = volume.split(':')[0].strip('- ')
    with open( '.' + host_path, 'w') as f:
        pass