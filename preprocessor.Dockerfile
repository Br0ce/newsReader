FROM golang:1.17-alpine AS build

WORKDIR /go/src/
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/app ./cmd/preprocessor/main.go

FROM alpine:latest
EXPOSE 8080
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/bin/app ./home

WORKDIR /home
RUN addgroup -S appgroup && adduser -S -D appuser -G appgroup
RUN chown appuser:appgroup ./app
USER appuser

ENTRYPOINT ["./app"]