FROM --platform=$BUILDPLATFORM golang:alpine AS build

RUN apk add --no-cache git

RUN git clone https://github.com/TheTipo01/xmasBot /xmasBot
WORKDIR /xmasBot
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o xmasBot

FROM alpine

RUN apk add --no-cache ffmpeg yt-dlp

COPY --from=build /xmasBot/xmasBot /usr/bin/

CMD ["xmasBot"]
