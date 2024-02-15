FROM golang:1.21.3-bullseye as build

WORKDIR /go/src

COPY . .
RUN go mod download
RUN go build -o ./app ./main.go
# Now copy it into our base image.
FROM gcr.io/distroless/base
COPY --from=build /go/src/app /go/src/app
# COPY --from=build /go/src/.env /go/src/.env

EXPOSE 8080

CMD ["/go/src/app"]