FROM base/archlinux:latest
MAINTAINER NPFLAN

RUN pacman --noconfirm -Syyu
RUN pacman -S --noconfirm kea python && \
mkdir -p /var/run/kea

ADD assets /etc/
ADD assets/kea.json /etc/kea.conf

EXPOSE 67/udp 67/tcp 68/udp 68/tcp

CMD kea-dhcp4 -c /etc/kea.conf
