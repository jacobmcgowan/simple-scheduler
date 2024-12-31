FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./shared/ ./shared/
COPY ./services/api/ ./services/api/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o api ./services/api/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=build /src/api ./
COPY --from=build /src/services/api/.env ./
EXPOSE 8080
ENTRYPOINT ["./api"]