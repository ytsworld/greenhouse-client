#! /bin/bash
set -e

# The build has to run on a raspberry device as cross compiling does not work for libraries that use "C"

cd /tmp/greenhouse-client

# Cleanup old binary if exists
if [ -e "./greenhouse-client" ]; then
    rm -f "./greenhouse-client"
fi

go get
go build -o greenhouse-client ./cmd

chmod +x greenhouse-client
echo Build was successful
