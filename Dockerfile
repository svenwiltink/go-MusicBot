FROM golang:1.6.4

run echo $GOPATH
RUN go get github.com/constabulary/gb/...
RUN useradd -ms /bin/bash musicbot

USER musicbot
WORKDIR /home/musicbot
