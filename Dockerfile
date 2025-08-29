FROM golang:alpine AS builder
RUN apk update && apk --no-cache add build-base
WORKDIR /go/src/github.com/NatoriMisong/livetv/
COPY . . 
RUN GO111MODULE=on CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o livetv .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata libc6-compat libgcc libstdc++ youtube-dl
WORKDIR /root
COPY --from=builder /go/src/github.com/NatoriMisong/livetv/view ./view
COPY --from=builder /go/src/github.com/NatoriMisong/livetv/assert ./assert
COPY --from=builder /go/src/github.com/NatoriMisong/livetv/.env .
COPY --from=builder /go/src/github.com/NatoriMisong/livetv/livetv .
EXPOSE 9000
VOLUME ["/root/data"]
CMD ["./livetv"]