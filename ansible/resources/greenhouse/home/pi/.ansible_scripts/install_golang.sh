#! /bin/bash
set -e

cd /tmp

# Remove previous installation
if [ -d "/usr/local/go" ]; then
    rm -rf /usr/local/go
fi

# Download and unzip go 1.12
wget https://dl.google.com/go/go1.12.9.linux-armv6l.tar.gz
tar -C /usr/local -xzf go1.12.9.linux-armv6l.tar.gz
rm -f go1.12.9.linux-armv6l.tar.gz

# symlink to bin dir
ln -s -f /usr/local/go/bin/go /usr/bin

# check installation
go version

touch /home/pi/.ansible_state/golang_installed