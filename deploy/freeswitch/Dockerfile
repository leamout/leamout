# syntax=docker/dockerfile:1

ARG FREESWITCH_REF=a88d069d6ffb74df797bcaf001f7e63181c07a09
ARG SOFIA_SIP_REF=ad36ac8
ARG FREESWITCH_BUILDER_IMAGE=signalwire/freeswitch-public-ci-base:bookworm-amd64
ARG FREESWITCH_RUNTIME_IMAGE=signalwire/freeswitch-public-base:bookworm

FROM ${FREESWITCH_BUILDER_IMAGE} AS build

ARG FREESWITCH_REF
ARG SOFIA_SIP_REF

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig

WORKDIR /usr/src

# FreeSWITCH 1.10.12 requires Sofia-SIP >= 1.13.17. Build the pinned
# 1.13.18 release locally so the image does not depend on an external
# FreeSWITCH binary package repository.
RUN git clone --depth 1 https://github.com/freeswitch/sofia-sip.git sofia-sip \
    && cd sofia-sip \
    && git fetch --depth 1 origin "${SOFIA_SIP_REF}" \
    && git checkout --detach "${SOFIA_SIP_REF}" \
    && ./bootstrap.sh \
    && ./configure --prefix=/usr/local --disable-rpath --disable-static --without-doxygen --disable-stun \
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

FROM ${FREESWITCH_RUNTIME_IMAGE} AS runtime

ENV PATH=/usr/local/freeswitch/bin:$PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/freeswitch/lib

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
