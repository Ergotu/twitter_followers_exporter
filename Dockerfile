FROM golang:alpine3.8 AS BUILD
WORKDIR /go/src/twitter_follower_extractor/
COPY *.go ./
RUN apk add git 
RUN go get -v
RUN go build

FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=BUILD /go/src/twitter_follower_extractor/twitter_follower_extractor ./
