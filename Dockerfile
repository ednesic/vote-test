FROM golang:alpine as builder
COPY . $GOPATH/src/Github.com/ednesic/vote-test/
WORKDIR $GOPATH/src/Github.com/ednesic/vote-test/
RUN go get -d -v && go build -o /go/bin/hello

FROM scratch
COPY --from=builder /go/bin/hello /go/bin/hello
ENTRYPOINT ["/go/bin/hello"]
