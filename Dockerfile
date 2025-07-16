FROM --platform=$BUILDPLATFORM golang:alpine AS build

RUN apk add --no-cache git
RUN apk add --no-cache wget
RUN apk add --no-cache nodejs
RUN apk add --no-cache pnpm

COPY . /yadmb

WORKDIR /yadmb/web
RUN pnpm install
RUN pnpm run build

WORKDIR /yadmb
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o yadmb

FROM alpine

RUN apk add --no-cache ffmpeg
RUN apk add --no-cache python3
RUN apk add --no-cache gcompat

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp && chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/

CMD ["yadmb"]
