FROM golang:alpine AS build

RUN apk add --no-cache \
  g++ \
  git \
  make

RUN git clone https://github.com/TheTipo01/YADMB /src
WORKDIR /src
RUN go mod download
RUN go build -trimpath github.com/bwmarrin/dca/cmd/dca
RUN strip dca
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o yadmb

FROM alpine

RUN apk add --no-cache \
  ffmpeg \
  opus \
  yt-dlp

COPY --from=build /src/yadmb /usr/bin/
COPY --from=build /src/dca /usr/bin/

CMD ["yadmb"]
