FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY . ./

RUN go build -o ./reverseProxy

EXPOSE 4105

CMD [ "./reverseProxy" ]