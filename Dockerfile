FROM --platform=$BUILDPLATFORM golang:alpine AS build

RUN apk add --no-cache git wget nodejs

RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.shrc" SHELL="$(which sh)" sh -

COPY . /yadmb

WORKDIR /yadmb/web
RUN . "$HOME/.shrc" && pnpm install --force
RUN . "$HOME/.shrc" && pnpm run build

WORKDIR /yadmb
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o yadmb

FROM alpine

RUN apk add --no-cache ffmpeg python3 gcompat

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp
RUN chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/

CMD ["yadmb"]
