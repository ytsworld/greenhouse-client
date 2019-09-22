#!/usr/bin/env bash

ps -ef | grep reducePower | grep -v grep | awk '{print $2}' | sudo xargs kill
