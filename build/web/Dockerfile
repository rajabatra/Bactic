# generate the api docs page

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
