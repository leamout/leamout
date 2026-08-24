# syntax=docker/dockerfile:1

ARG FREESWITCH_REF=a88d069d6ffb74df797bcaf001f7e63181c07a09
ARG FREESWITCH_BUILDER_IMAGE=signalwire/freeswitch-public-ci-base:bookworm-amd64
ARG FREESWITCH_RUNTIME_IMAGE=signalwire/freeswitch-public-base:bookworm

FROM ${FREESWITCH_BUILDER_IMAGE} AS build

ARG FREESWITCH_REF

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /usr/src

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

FROM ${FREESWITCH_RUNTIME_IMAGE} AS runtime

ENV PATH=/usr/local/freeswitch/bin:$PATH

COPY --from=build /usr/local/freeswitch /usr/local/freeswitch

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
