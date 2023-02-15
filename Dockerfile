FROM golang:1.20.1-bullseye AS build

WORKDIR /build

COPY go.* .

RUN go mod download

COPY . .

RUN go build  -ldflags="-w -s" -o /build/app cmd/app/main.go

FROM debian:bullseye

WORKDIR /app

COPY --from=build /build/app /app/app

ENTRYPOINT [ "/app/app" ]
