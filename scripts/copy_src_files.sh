#! /bin/bash
set -e

build_dir=/tmp/greenhouse-client

ssh pi@${PI_DEV_IP} "mkdir -p ${build_dir}"
scp go.* pi@${PI_DEV_IP}:${build_dir}
scp scripts/build.sh pi@${PI_DEV_IP}:${build_dir}
scp -r ./cmd pi@${PI_DEV_IP}:${build_dir}
scp -r ./pkg pi@${PI_DEV_IP}:${build_dir}
