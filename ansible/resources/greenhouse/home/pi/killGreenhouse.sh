#! /bin/bash

ps -ef | grep greenhouse-client | grep -v grep | awk '{print $2}' | sudo xargs kill -9
