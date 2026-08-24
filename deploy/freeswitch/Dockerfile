# syntax=docker/dockerfile:1

FROM debian:bookworm-slim

ARG FREESWITCH_PACKAGE_VERSION=1.10.12-1~bookworm+1

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
# Pin every FreeSWITCH package to the same Debian package release.
# This prevents a rebuild from silently pulling a newer FreeSWITCH.
# ------------------------------------------------------------

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    "freeswitch=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-commands=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-dptools=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-event-socket=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-sofia=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-console=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-logfile=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-dialplan-xml=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-codec-pcm=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-codec-g722=${FREESWITCH_PACKAGE_VERSION}" \
    "freeswitch-mod-codec-opus=${FREESWITCH_PACKAGE_VERSION}" \
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
