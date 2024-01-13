# generate the api docs page
# TODO: determine whether we need any of this apidocs generator stuff
#FROM openapitools/openapi-generator-cli:v7.0.1 as apidocs
#WORKDIR /usr/src/app
#COPY ./api/openapi.yaml .
#RUN bash /usr/local/bin/docker-entrypoint.sh generate -i openapi.yaml -g html2

FROM node:21-alpine as build
WORKDIR /usr/src/app
COPY ./web/package.json .
RUN npm install
COPY ./web .
RUN npm run build

FROM node:21-alpine as runtime
WORKDIR /srv
COPY --from=build /usr/src/app/package.json  ./
COPY --from=build /usr/src/app/build/ ./build 
COPY --from=build /usr/src/app/node_modules ./node_modules
CMD ["node", "build"]
