#!/bin/sh
firejail --profile=/home/ewe/.config/firejail/servex.profile --ip=192.168.8.101 --name=servex servex -addrs=':80'