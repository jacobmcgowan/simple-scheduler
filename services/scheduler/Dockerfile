FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./shared/ ./shared/
COPY ./services/scheduler/ ./services/scheduler/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o scheduler ./services/scheduler/main.go

FROM scratch
WORKDIR /app

COPY --from=build /src/scheduler ./
COPY --from=build /src/services/scheduler/docker.env ./.env
ENTRYPOINT ["./scheduler"]