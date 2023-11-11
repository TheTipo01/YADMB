FROM --platform=$BUILDPLATFORM golang:alpine AS build

RUN apk add --no-cache git wget nodejs

RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.shrc" SHELL="$(which sh)" sh -

RUN git clone https://github.com/TheTipo01/YADMB /yadmb

WORKDIR /yadmb/web
RUN . "$HOME/.shrc" && pnpm install
RUN . "$HOME/.shrc" && pnpm run build

WORKDIR /yadmb
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o yadmb

FROM thetipo01/dca

RUN apk add --no-cache ffmpeg python3 gcompat

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp
RUN chmod a+rx /usr/bin/yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/

CMD ["yadmb"]
