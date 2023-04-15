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

WORKDIR /

COPY --from=build /sheduler /sheduler

EXPOSE 8000

ENTRYPOINT ["/sheduler"]