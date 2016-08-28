FROM alpine:3.4

ARG ARCH
ADD bin/${ARCH}/spartakus /spartakus

RUN apk update --no-cache && apk add ca-certificates

ENTRYPOINT ["/spartakus"]
