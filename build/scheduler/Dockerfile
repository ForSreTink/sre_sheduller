##
## Build
##
FROM golang:1.20-buster AS build

WORKDIR /app

COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 go build -o /scheduler ./cmd/scheduler/main.go

##
## Deploy
##
FROM scratch

WORKDIR /app

COPY --from=build /app/openapi/api.yaml /app/openapi/api.yaml
COPY --from=build /app/internal/scheduler/server/templates/template.html /app/template.html
COPY --from=build /app/config.yml /app/config.yml
COPY --from=build /scheduler /app/scheduler

EXPOSE 8000

ENTRYPOINT ["/app/scheduler"]