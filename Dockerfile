FROM alpine:3.8
MAINTAINER NPFLAN

RUN apk add --no-cache \
 -X http://dl-4.alpinelinux.org/alpine/edge/testing \
 -X http://dl-4.alpinelinux.org/alpine/edge/main \
 kea kea-dhcp4 kea-keactrl kea-ctrl-agent

EXPOSE 67/udp 67/tcp 68/udp 68/tcp

ENTRYPOINT ["kea-dhcp4", "-c", "/etc/kea/kea-dhcp4.conf"]