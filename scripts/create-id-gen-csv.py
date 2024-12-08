import re
import os

with open('../docker-compose.yaml', 'r') as file:
    docker_compose_content = file.read()

volume_pattern = re.compile(r'- \./volumes/id-generator-[0-9].csv:/pkg/utils/id/id-generator-[0-9]')
volumes = volume_pattern.findall(docker_compose_content)

for volume in volumes:
    host_path = volume.split(':')[0].strip('- ')
    with open( '.' + host_path, 'w') as f:
        pass