FROM golang:1.24 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./shared/ ./shared/
COPY ./services/custodian/ ./services/custodian/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o custodian ./services/custodian/main.go

FROM scratch
WORKDIR /app

COPY --from=build /src/custodian ./
COPY --from=build /src/services/custodian/docker.env ./.env
ENTRYPOINT ["./custodian"]