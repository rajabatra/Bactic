# generate the api docs page

FROM node:21-alpine as dev
WORKDIR /usr/src/app
COPY ./package.json ./package-lock.json ./
RUN npm install
COPY . .

FROM node:21-alpine as build
WORKDIR /usr/src/app
COPY --from=dev /usr/src/app/ ./
RUN npm run build

FROM node:21-alpine as runtime
WORKDIR /srv
COPY --from=build /usr/src/app/package.json  ./
COPY --from=build /usr/src/app/build/ ./build 
COPY --from=build /usr/src/app/node_modules ./node_modules
CMD ["node", "build"]
