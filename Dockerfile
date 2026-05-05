FROM --platform=$BUILDPLATFORM golang:alpine AS build

COPY . /yadmb

WORKDIR /yadmb
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download

COPY --from=ghcr.io/thetipo01/yadmb-web:latest /web/build /yadmb/web/build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o yadmb

FROM alpine

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add ffmpeg python3 gcompat ca-certificates py3-pip && \
    pip3 install --break-system-packages "yt-dlp[default,curl-cffi]" yt-dlp-ejs && \
    apk del py3-pip

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/
COPY --from=denoland/deno:bin /deno /usr/bin/

COPY --from=build /yadmb/yadmb /usr/bin/

CMD ["yadmb"]