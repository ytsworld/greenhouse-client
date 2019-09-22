#!/bin/bash

#echo Pulling latest version of repo...
#git pull

#TODO Ugly, requried for docker desktop on windows
repo_sub_dir="F:\\Entwicklung\\git\\greenhouse-client"

if [ -z "$1" ]; then
  echo "Usage: $0 playbook-name.yml"
  exit 1
fi

docker run -it --rm -w /data -v ${repo_sub_dir}:/data teamidefix/ansible ansible-playbook -i inventory $1 -f 10
