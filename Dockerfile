FROM golang:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update
RUN apt-get -y install software-properties-common
RUN apt-add-repository -y "ppa:ubuntu-lxc/stable"
RUN apt-get -y update
RUN apt-get install git lxc lxc-template lxc-common lxc-dev libpcap-dev 