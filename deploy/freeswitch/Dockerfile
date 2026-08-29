# syntax=docker/dockerfile:1

# SignalWire's maintained public FreeSWITCH foundation already contains the
# packaged FreeSWITCH runtime and its native dependencies. Do not rebuild the
# dependency graph in this repository.
#
# The Bookworm image is pinned by its complete immutable manifest digest.
# Docker Hub reports this exact linux/amd64 manifest as:
# sha256:790ee70a1ae2725297ff00f5137acd93943267c242948af19799a9df325d767d
FROM signalwire/freeswitch-public-base:bookworm@sha256:790ee70a1ae2725297ff00f5137acd93943267c242948af19799a9df325d767d

USER root

# FreeSWITCH is installed under /usr/local/freeswitch in the maintained base
# image. Keep its CLI/runtime binaries on PATH so generic exec-style entrypoints
# and acceptance tooling can invoke freeswitch and fs_cli consistently.
ENV PATH="/usr/local/freeswitch/bin:${PATH}"

# Keep Leamout configuration in the image rather than mounting it as a
# mutable production volume. The existing configuration files themselves are
# intentionally unchanged.
COPY autoload_configs/ /leamout-config/autoload_configs/
COPY sip_profiles/ /leamout-config/sip_profiles/
COPY dialplan/ /leamout-config/dialplan/
COPY entrypoint.d/10-leamout-config.sh /docker-entrypoint.d/10-leamout-config.sh
COPY entrypoint.sh /usr/local/bin/leamout-freeswitch-entrypoint

RUN chmod 0755 \
      /docker-entrypoint.d/10-leamout-config.sh \
      /usr/local/bin/leamout-freeswitch-entrypoint \
    && mkdir -p /etc/freeswitch/autoload_configs \
      /etc/freeswitch/sip_profiles \
      /etc/freeswitch/dialplan \
    && command -v freeswitch >/dev/null \
    && command -v fs_cli >/dev/null

WORKDIR /etc/freeswitch

# SIP and RTP are intentionally documented here. Docker Compose decides which
# ports are actually published. ESL 8021 is deliberately NOT exposed/published.
EXPOSE 5060/udp
EXPOSE 5060/tcp
EXPOSE 16384-32768/udp

ENTRYPOINT ["/usr/local/bin/leamout-freeswitch-entrypoint"]
CMD ["freeswitch", "-nonat", "-nf"]
