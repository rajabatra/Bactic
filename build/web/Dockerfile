# generate the api docs page
FROM openapitools/openapi-generator-cli:v7.0.1 as apidocs
WORKDIR /usr/src/app
COPY ./api/openapi.yaml .
RUN bash /usr/local/bin/docker-entrypoint.sh generate -i openapi.yaml -g html2

FROM golang:1.21-alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

# copy over api docs generated in previous stage
COPY . .
COPY --from=apidocs /usr/src/app/index.html ./web/api.html

RUN go build -v -o /usr/local/bin/app ./cmd/site

CMD ["app"]