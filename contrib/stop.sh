#!/bin/bash
lxc-ls --frozen|grep -v honeytrap | while read line; do echo Stopping $line; lxc-unfreeze --name $line; done < /dev/stdin
lxc-ls --running|grep -v honeytrap | while read line; do echo Stopping $line; lxc-stop --name $line; done < /dev/stdin
lxc-ls --stopped|grep -v honeytrap | while read line; do echo Stopping $line; lxc-destroy --name $line; done < /dev/stdin
