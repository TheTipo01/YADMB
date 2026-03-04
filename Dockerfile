FROM --platform=$BUILDPLATFORM node:alpine AS web-build

RUN apk add --no-cache pnpm
COPY web /web

WORKDIR /web
RUN pnpm install
RUN pnpm run build

FROM golang:trixie AS build

RUN apt-get update && apt-get install build-essential unzip curl git make cmake -y

COPY . /yadmb

WORKDIR /yadmb
RUN go mod download

RUN wget https://raw.githubusercontent.com/disgoorg/godave/refs/heads/master/scripts/libdave_install.sh && chmod +x libdave_install.sh
ENV SHELL=/bin/sh
RUN ./libdave_install.sh v1.1.0

COPY --from=web-build /web/build /yadmb/web/build

ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"
RUN go build -trimpath -ldflags '-s -w' -o yadmb

FROM debian:trixie-slim

RUN apt-get update && apt-get install ffmpeg python3 curl ca-certificates unzip -y --no-install-recommends && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://deno.land/install.sh | sh

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/

RUN curl -o /usr/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]
