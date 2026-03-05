FROM --platform=$BUILDPLATFORM node:alpine AS web-build

RUN apk add --no-cache pnpm
COPY web /web

WORKDIR /web
RUN pnpm install
RUN pnpm run build

FROM golang:trixie AS build

RUN apt-get update && apt-get install unzip -y

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

RUN apt-get update && apt-get install ffmpeg python3 ca-certificates python3-pip -y --no-install-recommends && \
    pip3 install --break-system-packages --no-cache-dir "yt-dlp[default,curl-cffi]" yt-dlp-ejs && \
    apt-get purge -y --auto-remove python3-pip && rm -rf /var/lib/apt/lists/*

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/
COPY --from=denoland/deno:bin /deno /usr/bin/

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]
