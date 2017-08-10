FROM golang:1.8

WORKDIR /app
ENV SRC_DIR=/go/src/github.com/janisgarklavs/tournament/

ADD . $SRC_DIR
RUN cd $SRC_DIR; go get;  go build -o app; cp app /app/

ENTRYPOINT ["./app"]
