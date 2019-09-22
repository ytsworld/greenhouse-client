#!/usr/bin/env bash

timeToPowerSave=300


sleep $timeToPowerSave

systemctl stop ssh
/opt/vc/bin/tvservice -o
echo 0 > /sys/devices/platform/soc/3f980000.usb/buspower
echo 0 | sudo tee /sys/class/leds/led0/brightness 
echo 0 | sudo tee /sys/class/leds/led1/brightness
