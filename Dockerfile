FROM golang:1.17.2-alpine3.14

RUN addgroup -S -g 1000 yadmb \
  && adduser -S -G yadmb -u 999 yadmb \
  && mkdir /app \
  && chown yadmb:yadmb /app

ENV PYTHONUNBUFFERED=1

RUN apk add --no-cache \
  git \
  g++ \
  make \
  ffmpeg \
  youtube-dl \
  opus \
  python3 \
  py3-pip \
  && ln -sf python3 /usr/bin/python \
  && pip3 install lyricsgenius

USER yadmb

WORKDIR /app

RUN git clone https://github.com/TheTipo01/YADMB . \
  && go mod download \
  && go get -d github.com/bwmarrin/dca/cmd/dca \
  && go install github.com/bwmarrin/dca/cmd/dca

RUN go build -o build \
  && chmod +x ./build

CMD ["/app/build"]
