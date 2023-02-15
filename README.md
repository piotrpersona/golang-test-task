docker-compose up -d

Run golang apps:

## API

`go run cmd/api/main.go`

Addr: `:8080`

## Subscriber

`go run cmd/subscriber/main.go`


## Reporting API

`go run cmd/reporting/main.go`

Addr: `:8001`

`GET /message/list`
