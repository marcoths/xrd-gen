FROM golang:1.20-alpine AS build

RUN mkdir -p /out
RUN apk add --update --no-cache \
    bash \
    coreutils \
    git \
    make

RUN mkdir -p /go/src/github.com/jpmorganio-accelerator/batform/xgen
COPY . /go/src/github.com/jpmorganio-accelerator/batform/xgen/

RUN cd /go/src/github.com/jpmorganio-accelerator/batform/xgen && \
    make build

FROM alpine:3.18

WORKDIR /app

COPY --from=build /go/src/github.com/jpmorganio-accelerator/batform/xgen/bin/xgen /app/xgen
COPY --from=build /go/src/github.com/jpmorganio-accelerator/batform/xgen/hack /app/hack

RUN adduser -D user
RUN chown -R user:user /app
USER user:user

ENTRYPOINT ["/app/xgen"]