# syntax=docker/dockerfile:1

ARG FREESWITCH_REF=a88d069d6ffb74df797bcaf001f7e63181c07a09
ARG SOFIA_SIP_REF=v1.13.18

FROM debian:bookworm-slim AS build

ARG FREESWITCH_REF
ARG SOFIA_SIP_REF

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig

# Use only Debian's public repositories. Do not depend on SignalWire's
# package repositories for either FreeSWITCH or build dependencies.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      autoconf \
      automake \
      build-essential \
      ca-certificates \
      cmake \
      curl \
      git \
      libcurl4-openssl-dev \
      libedit-dev \
      libpcre3-dev \
      libpq-dev \
      libspeexdsp-dev \
      libsqlite3-dev \
      libssl-dev \
      libtool \
      libopus-dev \
      libsndfile1-dev \
      pkg-config \
      wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

# FreeSWITCH requires Sofia-SIP >= 1.13.17. Build the pinned v1.13.18 tag.
RUN git clone --depth 1 \
      --branch "${SOFIA_SIP_REF}" \
      https://github.com/freeswitch/sofia-sip.git \
      sofia-sip \
    && cd sofia-sip \
    && ./bootstrap.sh \
    && ./configure \
      --prefix=/usr/local \
      --disable-rpath \
      --disable-static \
      --without-doxygen \
      --disable-stun \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig

RUN git clone --depth 1 https://github.com/signalwire/freeswitch.git freeswitch \
    && cd freeswitch \
    && git fetch --depth 1 origin "${FREESWITCH_REF}" \
    && git checkout --detach "${FREESWITCH_REF}"

WORKDIR /usr/src/freeswitch

RUN ./bootstrap.sh -j

RUN ./configure \
      --disable-dependency-tracking \
      --enable-core-pgsql-support

RUN make -j"$(nproc)" \
    && make install

FROM debian:bookworm-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive
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
COPY --from=build /usr/local/lib/libsofia* /usr/local/lib/

RUN ldconfig

COPY autoload_configs/ /leamout-config/autoload_configs/
COPY sip_profiles/ /leamout-config/sip_profiles/
COPY dialplan/ /leamout-config/dialplan/

COPY entrypoint.d/10-leamout-config.sh \
    /docker-entrypoint.d/10-leamout-config.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /docker-entrypoint.d/10-leamout-config.sh /entrypoint.sh \
    && mkdir -p /usr/local/freeswitch/recordings

WORKDIR /usr/local/freeswitch

EXPOSE 5060/udp
EXPOSE 5060/tcp
EXPOSE 16384-32768/udp

ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
