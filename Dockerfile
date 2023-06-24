FROM golang:alpine as builder
WORKDIR /app
ARG ldflags
RUN apk update && apk upgrade && apk add --no-cache ca-certificates && apk add --no-cache build-base  && apk add --no-cache git
RUN update-ca-certificates
COPY . .
# go1.18beta2 build -ldflags="${ldflags}" -o dist/skiver${SUFFIX} main.go
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="${ldflags}" -o gotally ./api/cmd/

FROM alpine as scratch

WORKDIR /app

COPY --from=builder /app/gotally .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN mkdir data
ENTRYPOINT [ "/app/gotally" ]
