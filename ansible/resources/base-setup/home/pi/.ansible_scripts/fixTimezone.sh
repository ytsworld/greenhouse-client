#!/bin/bash

sudo rm /etc/localtime
sudo ln -s /usr/share/zoneinfo/Europe/Berlin /etc/localtime
sudo rm /etc/timezone
echo "Europe/Berlin" | sudo tee /etc/timezone

touch /home/pi/.ansible_state/timezoneisalreadyset
