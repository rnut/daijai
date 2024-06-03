FROM golang:1.21.3-bullseye as build

WORKDIR /go/src

COPY . .
COPY .env.example.prd .env
COPY keys/daijai-d4ab4aa6981d.json keys/daijai-d4ab4aa6981d.json
RUN go mod download
RUN go build -o ./app ./main.go


FROM gcr.io/distroless/base-debian11
COPY --from=build /go/src/keys/daijai-d4ab4aa6981d.json /keys/daijai-d4ab4aa6981d.json
COPY --from=build /go/src/.env /
COPY --from=build /go/src/app /

EXPOSE 8080

CMD ["/app"]