FROM --platform=$BUILDPLATFORM node:alpine AS web-build

RUN apk add --no-cache pnpm
COPY web /web

WORKDIR /web
RUN pnpm install
RUN pnpm run build

FROM ghcr.io/thetipo01/godave-musl:latest AS build

COPY . /yadmb

WORKDIR /yadmb
RUN go mod download

COPY --from=web-build /web/build /yadmb/web/build

RUN go build -trimpath -ldflags "-s -w" -o yadmb

FROM alpine

RUN apk add --no-cache ffmpeg python3 gcompat deno

COPY --from=ghcr.io/thetipo01/dca:latest /usr/bin/dca /usr/bin/

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp && chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]
