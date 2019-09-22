#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 playbook-name.yml"
  exit 1
fi

sudo docker run -it --rm -w /data -v `pwd`:/data teamidefix/ansible ansible-playbook -i inventory $1 -f 1 -vvv
