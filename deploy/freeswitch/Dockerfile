# syntax=docker/dockerfile:1

# Build FreeSWITCH with the official project's dependency/package pipeline.
# This avoids treating the CI base as if it were a preinstalled Sofia-SIP /
# SpanDSP development environment.
FROM debian:bookworm AS build

ARG FREESWITCH_REF=3f13ad1b1d323e435757fc43a2bb829d4ea5b9a8

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      ca-certificates \
      git \
      lsb-release \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

# Pin the complete FreeSWITCH source tree to the v1.11.2 release commit.
RUN git clone --depth 1 https://github.com/signalwire/freeswitch.git freeswitch \
    && cd freeswitch \
    && git fetch --depth 1 origin "${FREESWITCH_REF}" \
    && git checkout --detach "${FREESWITCH_REF}"

WORKDIR /usr/src/freeswitch

# FreeSWITCH's own packaging dependency builder is the source of truth for
# Sofia-SIP, SpanDSP, and the rest of the native dependency graph.
# It produces Debian packages rather than requiring this image to guess which
# development libraries configure needs.
FROM signalwire/freeswitch-public-ci-base:bookworm-amd64 AS toolchain

# Keep this stage separate: the official CI image supplies the project's build
# tooling, while the dependency/package scripts remain owned by FreeSWITCH.
COPY --from=build /usr/src/freeswitch /usr/src/freeswitch

WORKDIR /usr/src/freeswitch

RUN ./docker/build/debs-from-source.Dockerfile
