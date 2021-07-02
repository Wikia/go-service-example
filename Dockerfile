FROM gcr.io/distroless/static-debian10

COPY go-example-service /svc/
WORKDIR /svc

USER 65534:65534

EXPOSE 3000:3000
EXPOSE 4000:4000
EXPOSE 5000:5000

ENTRYPOINT ["/svc/go-example-service"]