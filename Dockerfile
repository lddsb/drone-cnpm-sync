# builder
FROM golang:latest AS builder

WORKDIR /go/src/github.com/lddsb/drone-cnpm-sync

COPY . .

ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# real base image
FROM alpine:latest AS publish

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/lddsb/drone-cnpm-sync/app /bin/

ENTRYPOINT ["/bin/app"]
