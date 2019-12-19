FROM golang:1.11-alpine

WORKDIR /go/src/github.com/astravexton/irchuu
COPY . .

RUN adduser -s /bin/bash -g 1002 -u 1002 -D -H irchuu && \
    mkdir -p /go/irchuu/data && \
    chown -R 1002:1002 /go/irchuu

RUN apk update && \
    apk add --no-cache git bash && \
    go get -d -v ./... && \
    cd cmd/irchuu && \
    go install && \
    apk del git && \
    rm -fr /go/src/github.com/astravexton/irchuu

USER irchuu

CMD ["irchuu", "-config", "/go/irchuu/irchuu.conf", "-data", "/go/irchuu/data/"]
