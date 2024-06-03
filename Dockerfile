# Build
FROM    golang:latest AS build

ENV     GO111MODULE=on

RUN     mkdir /go/src/app
WORKDIR /go/src/app
ADD     . /go/src/app
RUN     CGO_ENABLED=0 make deps
RUN     CGO_ENABLED=0 make build


# Deploy
FROM    alpine:latest
COPY    --from=build /go/src/app/bin/sdi ./
EXPOSE  8080
CMD     ["./sdi"]
