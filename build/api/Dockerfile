FROM golang:1.22 AS build
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0
RUN go build -v -o ./api ./cmd/api

FROM scratch AS runtime
WORKDIR /usr/bin/
COPY --from=build /go/src/api ./
EXPOSE 8080/tcp
ENTRYPOINT ["api"]
