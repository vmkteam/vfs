FROM golang:1.18-alpine@sha256:77f25981bd57e60a510165f3be89c901aec90453fd0f1c5a45691f6cb1528807 AS builder
ADD . /build
RUN cd /build && go install -mod=mod ./cmd/vfssrv

FROM alpine:latest

ENV TZ=Europe/Moscow
RUN apk --no-cache add ca-certificates tzdata && cp -r -f /usr/share/zoneinfo/$TZ /etc/localtime

COPY --from=builder /go/bin/vfssrv .

ENTRYPOINT ["/vfssrv"]
EXPOSE 9999
