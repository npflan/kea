FROM base/archlinux
MAINTAINER NPFLAN

RUN pacman -Syyu
RUN pacman --noconfirm -S kea python

ADD assets/kea.json /etc/kea.conf
ADD assets/kea-ca.conf /etc/kea-ca.conf
ADD assets/keactrl.conf /etc/keactrl.conf

ADD assets/entry-point.sh /entry-point.sh
RUN chmod +x /entry-point.sh
RUN mkdir -p /var/run/kea

EXPOSE 67/udp 67/tcp 68/udp 68/tcp

CMD kea-dhcp4 -c /etc/kea.conf
