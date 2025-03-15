FROM golang:1.24-alpine AS build

WORKDIR /code

COPY . .
RUN go build -o app cmd/cmd.go

FROM alpine:3.18

WORKDIR /app
COPY --from=build /code/app .

ENTRYPOINT ["/app/app"]
