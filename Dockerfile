FROM gcr.io/google-containers/ubuntu-slim:0.14
MAINTAINER NPFLAN

<<<<<<< HEAD
RUN apt-get update
RUN apt-get install kea-dhcp4-server python && \
mkdir -p /var/run/kea

ADD assets /etc/
ADD assets/kea.json /etc/kea.conf

EXPOSE 67/udp 67/tcp 68/udp 68/tcp

CMD kea-dhcp4 -c /etc/kea.conf
