FROM golang:alpine as builder
RUN apk add git build-base
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN make build-alpine

FROM scratch
COPY --from=builder /build/bin/go-service-example /app/
WORKDIR /app

USER 65534:65534

EXPOSE 3000:3000
EXPOSE 4000:4000
EXPOSE 5000:5000

ENTRYPOINT ["/app/go-service-example"]