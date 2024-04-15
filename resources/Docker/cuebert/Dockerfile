FROM golang:1.21.1 as builder

WORKDIR /go/src/github.com/johnmikee/cuebert

ENV CGO_ENABLED=1 \
	GOOS=linux

COPY . .

RUN make tidy
RUN make cuebert

FROM builder

COPY --from=builder /go/src/github.com/johnmikee/cuebert/build/linux/cuebert /usr/bin/cuebert

CMD ["/usr/bin/cuebert"]
