FROM golang:1.24 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./shared/ ./shared/
COPY ./services/api/ ./services/api/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o api ./services/api/main.go

FROM scratch
WORKDIR /app

COPY --from=build /src/api ./
COPY --from=build /src/services/api/docker.env ./.env
EXPOSE 8080
ENTRYPOINT ["./api"]