FROM golang:1.20.7-alpine

RUN addgroup -S -g 1000 yadmb \
  && adduser -S -G yadmb -u 999 yadmb \
  && mkdir /app \
  && chown yadmb:yadmb /app

RUN apk add --no-cache \
  git \
  g++ \
  make \
  ffmpeg \
  opus \
  yt-dlp

USER yadmb

WORKDIR /app

RUN git clone https://github.com/TheTipo01/YADMB . \
  && go mod download \
  && go get -d github.com/bwmarrin/dca/cmd/dca \
  && go install github.com/bwmarrin/dca/cmd/dca

RUN go build -o build \
  && chmod +x ./build

CMD ["/app/build"]
