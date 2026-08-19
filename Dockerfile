FROM golang:alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add --no-cache build-base pkgconfig ccache

COPY go.mod /yadmb/go.mod
COPY go.sum /yadmb/go.sum

WORKDIR /yadmb

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY --from=ghcr.io/thetipo01/yadmb-web:latest /web/build /yadmb/web/build

COPY . /yadmb

COPY --from=ghcr.io/thetipo01/godave-musl:latest /root/.local /root/.local
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

ENV CC=/usr/local/bin/gcc CXX=/usr/local/bin/g++
RUN ln -s /usr/bin/ccache /usr/local/bin/gcc && ln -s /usr/bin/ccache /usr/local/bin/g++ && ln -s /usr/bin/ccache /usr/local/bin/cc && ln -s /usr/bin/ccache /usr/local/bin/c++
ENV CCACHE_DIR=/ccache

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/ccache \
    go build -trimpath -ldflags '-s -w' -o yadmb

FROM alpine

RUN --mount=type=cache,target=/var/cache/apk \
    --mount=type=cache,target=/root/.cache/pip \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add ffmpeg python3 ca-certificates py3-pip && \
    pip3 install --break-system-packages --pre "yt-dlp[default,curl-cffi]" yt-dlp-ejs && \
    apk del py3-pip

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/
COPY --from=denoland/deno:bin /deno /usr/bin/

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]