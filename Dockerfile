FROM golang:bookworm AS build

RUN apt update && apt install git build-essential libopus-dev autoconf libtool pkg-config -y

WORKDIR /opus
RUN git clone https://github.com/xiph/opus /opus
RUN git checkout 1.1.2
RUN ./autogen.sh
RUN ./configure
RUN make
ARG PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:/opus"

RUN git clone https://github.com/TheTipo01/YADMB /yadmb
WORKDIR /yadmb
RUN go build -trimpath -ldflags "-s -w" -o yadmb
RUN go install github.com/bwmarrin/dca/cmd/dca@latest
RUN strip /go/bin/dca

FROM alpine

RUN apk add --no-cache \
  ffmpeg \
  yt-dlp

COPY --from=build /yadmb/yadmb /usr/bin/
COPY --from=build /go/bin/dca /usr/bin/

CMD ["yadmb"]
