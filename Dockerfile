FROM docker.io/library/golang:1.14-alpine as builder

LABEL MAINTAINER="Jack Murdock <jack_murdock@comcast.com>"

WORKDIR /go/src/github.com/xmidt-org/go-parodus

RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    git \
    openssh \
    gcc \
    libc-dev \
    upx \
    cmake autoconf make musl-dev gcc g++ openssl openssl-dev git cunit cunit-dev automake libtool util-linux-dev

COPY . .
RUN make build
RUN CGO_ENABLED=0 go build  -a -ldflags "-w -s"  -o req-res github.com/xmidt-org/go-parodus/examples/request-response/

FROM simulator:latest

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/xmidt-org/go-parodus/parodus /go/src/github.com/xmidt-org/go-parodus/req-res /
COPY entrypoint.sh /

USER nobody

ENTRYPOINT ["/entrypoint.sh"]
