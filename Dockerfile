FROM ghcr.io/thetipo01/godave-musl:latest AS build

RUN apk add --no-cache git wget nodejs pnpm

COPY . /yadmb

WORKDIR /yadmb/web
RUN pnpm install
RUN pnpm run build

WORKDIR /yadmb
RUN go mod download
RUN go build -trimpath -ldflags "-s -w" -o yadmb

FROM alpine

RUN apk add --no-cache ffmpeg python3 gcompat deno

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp && chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /root/.local/lib/pkgconfig /root/.local/lib/pkgconfig
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["yadmb"]
