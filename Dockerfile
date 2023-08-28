FROM golang:alpine AS build

RUN apk add --no-cache \
  g++ \
  git \
  make \
  opus \
  pkgconfig

RUN git clone https://github.com/TheTipo01/YADMB /src
WORKDIR /src
RUN go build -trimpath -ldflags "-s -w" -o yadmb
RUN go install github.com/bwmarrin/dca/cmd/dca@latest
RUN strip /go/bin/dca

FROM alpine

RUN apk add --no-cache \
  ffmpeg \
  yt-dlp

COPY --from=build /src/yadmb /usr/bin/
COPY --from=build /go/bin/dca /usr/bin/

CMD ["yadmb"]
