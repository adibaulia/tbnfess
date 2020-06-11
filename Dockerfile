## Builder
FROM golang:1.13.0-alpine3.10 as builder

LABEL name="tbnfess"
LABEL version="1.0.0"

RUN apk update

RUN mkdir -p /home/go/app

WORKDIR /home/go/app

ENV CGO_ENABLED=0
ENV GOOS=linux

COPY . .
RUN go mod tidy
RUN go mod vendor
EXPOSE 80
RUN chmod +x /home/go/app/entrypoint.sh
CMD ["/bin/sh", "/home/go/app/entrypoint.sh"]
