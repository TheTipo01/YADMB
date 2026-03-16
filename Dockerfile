# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:alpine AS web-build

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add pnpm
COPY web /web

ENV PNPM_HOME="/pnpm"

WORKDIR /web
RUN --mount=type=cache,target=${PNPM_HOME} \
    pnpm config set store-dir ${PNPM_HOME} && \
    pnpm install --frozen-lockfile --prefer-offline && \
    pnpm install
RUN pnpm run build

FROM golang:trixie AS build

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && apt-get install unzip -y

COPY . /yadmb

WORKDIR /yadmb
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN wget https://raw.githubusercontent.com/disgoorg/godave/refs/heads/master/scripts/libdave_install.sh && chmod +x libdave_install.sh
ENV SHELL=/bin/sh
RUN ./libdave_install.sh v1.1.0

COPY --from=web-build /web/build /yadmb/web/build

ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags '-s -w' -o yadmb

FROM debian:trixie-slim

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    --mount=type=cache,target=/root/.cache/pip \
    apt-get update && apt-get install ffmpeg python3 ca-certificates python3-pip -y --no-install-recommends && \
    pip3 install --break-system-packages "yt-dlp[default,curl-cffi]" yt-dlp-ejs && \
    apt-get purge -y --auto-remove python3-pip

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/
COPY --from=denoland/deno:bin /deno /usr/bin/

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]
