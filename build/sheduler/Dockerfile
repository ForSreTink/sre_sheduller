##
## Build
##
FROM golang:1.20-buster AS build

WORKDIR /app

COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 go build -o /sheduler ./cmd/sheduler/main.go

##
## Deploy
##
FROM scratch

WORKDIR /app

COPY --from=build /app/openapi/api.yaml /app/openapi/api.yaml
COPY --from=build /sheduler /app/sheduler

EXPOSE 8000

ENTRYPOINT ["/app/sheduler"]