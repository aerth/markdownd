FROM golang:latest as builder
MAINTAINER aerth <aerth@riseup.net>
WORKDIR /src
COPY . .
RUN go build -o /src/markdownd -tags "usergo,netgo" -v -ldflags='-w -s'
RUN cp /src/markdownd /bin/markdownd
RUN rm -rf /src

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /bin/markdownd /bin/markdownd

EXPOSE 8080
ENTRYPOINT ["markdownd"]
