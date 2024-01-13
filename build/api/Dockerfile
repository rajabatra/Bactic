FROM golang:1.19 AS build
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o ./api ./cmd/api

EXPOSE 8080/tcp

ENV CGO_ENABLED=0

RUN go build -o openapi .

FROM scratch AS runtime
COPY --from=build /go/src/api ./
EXPOSE 8080/tcp
ENTRYPOINT ["./api"]
