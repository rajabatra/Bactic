FROM golang:1.22 AS build
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0
RUN go build -v -o ./scraper ./cmd/scraper

FROM scratch AS runtime
WORKDIR /usr/bin
COPY --from=build /go/src/scraper ./
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["scraper"]
