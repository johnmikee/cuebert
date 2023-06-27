FROM golang:1.19 as builder

WORKDIR /go/src/github.com/johnmikee/cuebert

ARG TARGETARCH
ARG TARGETOS
ARG SENTRY_RELEASE=dev
ARG DD_VERSION=dev

ENV CGO_ENABLED=0 \
	GOARCH=$TARGETARCH \
	GOOS=$TARGETOS \
	DD_VERSION=$DD_VERSION \
	SENTRY_RELEASE=$SENTRY_RELEASE

COPY . .

RUN make deps
RUN make tidy
RUN make cuebert

FROM alpine:latest

COPY --from=builder /go/src/github.com/johnmikee/cuebert/build/linux/cuebert /usr/bin/cuebert

CMD ["/usr/bin/cuebert"]
