# Build
FROM    golang:latest AS build

ENV     GO111MODULE=on

RUN     mkdir /go/src/app
WORKDIR /go/src/app
ADD     . /go/src/app
RUN     CGO_ENABLED=0 make deps
RUN     CGO_ENABLED=0 make build
EXPOSE  8080 4000


# Deploy
FROM    alpine:latest
RUN     mkdir /opt/sportix
WORKDIR /opt/sportix
COPY    --from=build /go/src/app/bin/sdi /opt/sportix/sportix
COPY    --from=build /go/src/app/public /opt/sportix/public
EXPOSE  8080
CMD     ["./sportix", "--dir", "./pb_data", "serve", "--http=0.0.0.0:8080"]
