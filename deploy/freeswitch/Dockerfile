# syntax=docker/dockerfile:1

# FreeSWITCH's maintained CI foundation supplies the build toolchain. We do
# not consume its authenticated SignalWire package repository; all native
# FreeSWITCH dependencies are built by the pinned source tree below.
FROM signalwire/freeswitch-public-ci-base:bookworm-amd64 AS build

ARG FREESWITCH_REF=3f13ad1b1d323e435757fc43a2bb829d4ea5b9a8
ARG DEP_BUILD_NUMBER=42

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

# The CI foundation can contain SignalWire APT sources using either .list or
# deb822 .sources files. Remove every source file that references SignalWire's
# authenticated repositories. Do not rely on a filename convention.
RUN while IFS= read -r -d '' file; do \
      if grep -Eq '(^|[[:space:]]|=)(https?://)?(packages|freeswitch)\.signalwire\.com/' "$file"; then \
        echo "Removing authenticated SignalWire APT source: $file"; \
        rm -f "$file"; \
      fi; \
    done < <(find /etc/apt -type f -print0) \
    && rm -f /etc/apt/auth.conf /etc/apt/auth.conf.d/* 2>/dev/null || true

WORKDIR /usr/src

# Pin FreeSWITCH to the v1.11.2 release commit.
RUN git clone --depth 1 https://github.com/signalwire/freeswitch.git freeswitch \
    && cd freeswitch \
    && git fetch --depth 1 origin "${FREESWITCH_REF}" \
    && git checkout --detach "${FREESWITCH_REF}"

WORKDIR /usr/src/freeswitch

# Use FreeSWITCH's maintained dependency definitions. The builder clones and
# builds Sofia-SIP, SpanDSP, libks, signalwire-c, and the other native
# dependencies into /var/local/deb. --setup is intentionally omitted because
# this image already has its build toolchain and must not install SignalWire's
# authenticated package infrastructure.
RUN ./scripts/packaging/build/dependencies/build-dependencies.sh \
      --build-number "${DEP_BUILD_NUMBER}" \
      --all \
      --repo \
      --clone \
      --git-https

# The dependency builder creates a local file:// repository in /var/local/deb.
# Install the generated packages and their dependencies.
RUN apt-get update \
    && find /var/local/deb -maxdepth 1 -type f -name '*.deb' -print0 \
      | xargs -0 -r apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

RUN ./bootstrap.sh -j \
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

# ESL 8021 remains private and is intentionally not published by this image.
ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
