FROM golang:1.13 as builder

ARG APP_NAME=example_app

RUN mkdir /gocache
ENV GOCACHE /gocache

WORKDIR /go/src/service
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o app cmd/${APP_NAME}/main.go

FROM scratch

COPY --from=builder /go/src/service/app /service/app

WORKDIR /service

EXPOSE 3000, 4000

ENTRYPOINT ["/service/app"]