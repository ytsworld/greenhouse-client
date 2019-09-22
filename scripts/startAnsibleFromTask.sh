#!/bin/bash

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: $0 playbook-name.yml \"task name to start from\""
  exit 1
fi

docker run -it --rm -w /data -v `pwd`:/data teamidefix/ansible ansible-playbook -i inventory $1 --start-at-task="$2"
