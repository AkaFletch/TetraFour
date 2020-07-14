FROM golang:alpine
WORKDIR /go/src/app
COPY . .

RUN go get -u -d -v ./...
RUN go install -v ./...
RUN go build -v .
RUN mv config.yml.docker config.yml

CMD ["tetrafour"]
