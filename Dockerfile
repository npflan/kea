FROM alpine:latest
MAINTAINER NPFLAN

RUN apk update && \
    apk add procps alpine-sdk git autoconf automake libressl libressl-dev boost-dev libtool pkgconfig postgresql postgresql-dev && \
    cd /tmp && \
    git clone -b 1.2.x https://github.com/log4cplus/log4cplus.git && \
    cd log4cplus && \
    git submodule update --init --recursive && \
    autoreconf && \
    ./configure && \
    make && \
    make install && \
    cd /tmp && \
    git clone https://github.com/isc-projects/kea.git && \
    cd kea && \
    autoreconf --install && \
    ./configure --with-dhcp-pgsql && \
    make && \
    make install && \
    rm -rf /tmp/* && \
    apk del alpine-sdk git autoconf automake pkgconfig && \
    rm -rf /var/cache/apk/*

RUN mkdir -p /var/run/kea

ADD assets /etc/
ADD assets/kea.json /etc/kea.conf

EXPOSE 67/udp 67/tcp 68/udp 68/tcp

CMD kea-dhcp4 -c /etc/kea.conf
