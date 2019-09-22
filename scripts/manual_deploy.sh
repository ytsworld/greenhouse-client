#! /bin/bash
set -e

ssh pi@${PI_PROD_IP} '/bin/bash -c "ps -ef " | grep greenhouse-client | grep -v grep | awk "{print \$2}" | sudo xargs kill -9'
scp ./greenhouse-client pi@${PI_PROD_IP}:/home/pi
