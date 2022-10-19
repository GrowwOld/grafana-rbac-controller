FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /grafana-rbac-controller

EXPOSE 9080

CMD [ "/grafana-rbac-controller" ]