# syntax=docker/dockerfile:1

FROM debian:bookworm-slim

ARG FREESWITCH_VERSION=1.10.12

ENV DEBIAN_FRONTEND=noninteractive

# ------------------------------------------------------------
# Base dependencies
# ------------------------------------------------------------

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    procps \
    tini \
    && rm -rf /var/lib/apt/lists/*

# ------------------------------------------------------------
# FreeSWITCH repository
# ------------------------------------------------------------

RUN mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://packages.signalwire.com/repo/deb/debian-release/signalwire-unstable.gpg \
    -o /etc/apt/keyrings/signalwire.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/signalwire.gpg] https://packages.signalwire.com/repo/deb/debian-release/ bookworm main" \
    > /etc/apt/sources.list.d/freeswitch.list

# ------------------------------------------------------------
# FreeSWITCH
# ------------------------------------------------------------

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    freeswitch \
    freeswitch-mod-commands \
    freeswitch-mod-dptools \
    freeswitch-mod-event-socket \
    freeswitch-mod-sofia \
    freeswitch-mod-console \
    freeswitch-mod-logfile \
    freeswitch-mod-dialplan-xml \
    freeswitch-mod-codec-pcm \
    freeswitch-mod-codec-g722 \
    freeswitch-mod-codec-opus \
    && rm -rf /var/lib/apt/lists/*

# ------------------------------------------------------------
# Leamout configuration
# ------------------------------------------------------------

COPY autoload_configs/ /leamout-config/autoload_configs/
COPY sip_profiles/ /leamout-config/sip_profiles/
COPY dialplan/ /leamout-config/dialplan/

COPY entrypoint.d/10-leamout-config.sh \
    /docker-entrypoint.d/10-leamout-config.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /docker-entrypoint.d/10-leamout-config.sh /entrypoint.sh

# ------------------------------------------------------------
# FreeSWITCH runtime
# ------------------------------------------------------------

WORKDIR /var/lib/freeswitch

EXPOSE 5060/udp
EXPOSE 5060/tcp

# RTP range used by the internal SIP profile.
EXPOSE 16384-32768/udp

ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
