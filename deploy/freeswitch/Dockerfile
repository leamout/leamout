# syntax=docker/dockerfile:1

# FreeSWITCH's maintained Debian Bookworm CI/build foundation.
# Keep the tag here until the registry's complete immutable digest has been
# independently verified; never invent or truncate a digest.
FROM signalwire/freeswitch-public-ci-base:bookworm-amd64 AS build

# v1.11.2 was released on 2026-08-09. Pin the source to the release commit,
# rather than relying on a moving branch or tag at build time.
ARG FREESWITCH_REF=3f13ad1b1d323e435757fc43a2bb829d4ea5b9a8

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /usr/src

# Use the official FreeSWITCH source and its maintained CI build environment.
# The CI base already provides the build toolchain/dependencies; do not add
# SignalWire APT repositories or manually build Sofia-SIP/SpanDSP here.
RUN git clone --depth 1 \
      https://github.com/signalwire/freeswitch.git \
      freeswitch \
    && cd freeswitch \
    && git fetch --depth 1 origin "${FREESWITCH_REF}" \
    && git checkout --detach "${FREESWITCH_REF}" \
    && ./bootstrap.sh -j \
    && ./configure \
      --disable-dependency-tracking \
      --enable-core-pgsql-support \
    && make -j"$(nproc)" \
    && make install

# Keep the runtime as a Leamout-owned image. It deliberately does not expose
# FreeSWITCH ESL (8021); only the SIP and RTP listeners are part of the image's
# public network contract.
FROM debian:bookworm-slim AS runtime

ENV PATH=/usr/local/freeswitch/bin:$PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/freeswitch/lib

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      ca-certificates \
      libcurl4 \
      libedit2 \
      libpcre3 \
      libpq5 \
      libspeexdsp1 \
      libsqlite3-0 \
      libssl3 \
      libopus0 \
      libsndfile1 \
      tini \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/freeswitch /usr/local/freeswitch
COPY --from=build /usr/local/lib/ /usr/local/lib/

RUN ldconfig

# Existing Leamout SIP/RTP configuration remains unchanged.
COPY autoload_configs/ /leamout-config/autoload_configs/
COPY sip_profiles/ /leamout-config/sip_profiles/
COPY dialplan/ /leamout-config/dialplan/
COPY entrypoint.d/10-leamout-config.sh \
    /docker-entrypoint.d/10-leamout-config.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x \
      /docker-entrypoint.d/10-leamout-config.sh \
      /entrypoint.sh \
    && mkdir -p /usr/local/freeswitch/recordings

WORKDIR /usr/local/freeswitch

EXPOSE 5060/udp
EXPOSE 5060/tcp
EXPOSE 16384-32768/udp

ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
