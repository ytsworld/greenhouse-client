#!/usr/bin/env bash

timeToPowerSave=300


sleep $timeToPowerSave

/opt/vc/bin/tvservice -o
echo 0 | sudo tee /sys/class/leds/led0/brightness 
echo 0 | sudo tee /sys/class/leds/led1/brightness

# Turning off ssh daemon does not save energy but might be a good idea on production devices
# systemctl stop ssh

# TODO: Needs analysis - this seems to break the client
# echo 0 > /sys/devices/platform/soc/3f980000.usb/buspower
