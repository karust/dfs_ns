FROM golang:1.11.1-alpine3.8
RUN mkdir -p /Server/
COPY . /Server/
RUN cd /Server/ && \
    apk update && \
    apk upgrade && \
    apk add sqlite && \
    apk add git && \
    apk add build-base && \
    go get "github.com/mattn/go-sqlite3" && \
    go get "github.com/jinzhu/gorm" && \
    go get "github.com/dgrijalva/jwt-go" && \
    go get "github.com/julienschmidt/httprouter" && \
    go get "github.com/rs/cors" && \
    go build && \
    chmod 777 ./Server

WORKDIR /Server/

CMD ["./Server"]