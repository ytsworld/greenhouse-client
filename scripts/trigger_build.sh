#! /bin/bash
set -e

build_dir=/tmp/greenhouse-client

./scripts/copy_src_files.sh
ssh pi@${PI_DEV_IP} ${build_dir}/build.sh
scp pi@${PI_DEV_IP}:${build_dir}/greenhouse-client .
