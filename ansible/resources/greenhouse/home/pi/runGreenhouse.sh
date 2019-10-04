#! /bin/sh

GOOGLE_APPLICATION_CREDENTIALS="/home/pi/greenhouse-client.sa.json"
export GOOGLE_APPLICATION_CREDENTIALS
GREENHOUSE_RECEIVER_URL="{{ greenhouse_recv_url }}"
export GREENHOUSE_RECEIVER_URL

/home/pi/greenhouse-client
