FROM golang:alpine as builder
RUN apk add --no-cache git mercurial
COPY . $GOPATH/src/Github.com/ednesic/vote-test/
WORKDIR $GOPATH/src/Github.com/ednesic/vote-test/
RUN go get -d -v \ 
    && go build -o /go/bin/app

FROM alpine
COPY --from=builder /go/bin/app /
ENTRYPOINT ["/app"]
