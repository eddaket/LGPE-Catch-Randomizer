FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/app/main.go

###################

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/app .
COPY --from=build /app/data ./data
COPY --from=build /app/templates ./templates

EXPOSE 8080

CMD ["./app"]
