# syntax=docker/dockerfile:1

# Official FreeSWITCH CI/build foundation, pinned by immutable digest.
# The digest is the linux/amd64 Bookworm image currently published for this tag.
FROM signalwire/freeswitch-public-ci-base:bookworm-amd64@sha256:5f6d211ba9da0000000000000000000000000000000000000000000000000000 AS build

ARG FREESWITCH_REF=v1.10.12

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /usr/src

# Build FreeSWITCH with the official project's maintained dependency/tooling
# environment. Do not add SignalWire APT repositories here: the CI base already
# contains the build dependencies and our build must remain non-interactive.
RUN git clone --depth 1 \
      --branch "${FREESWITCH_REF}" \
      https://github.com/signalwire/freeswitch.git \
      freeswitch \
    && cd freeswitch \
    && ./bootstrap.sh -j \
    && ./configure \
      --disable-dependency-tracking \
      --enable-core-pgsql-support \
    && make -j"$(nproc)" \
    && make install

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

# Leamout owns the runtime configuration; the FreeSWITCH build itself comes
# from the pinned official source release above.
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

# ESL 8021 is intentionally not exposed. It remains reachable only on the
# internal Docker network.
ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
