FROM node:16-alpine AS build

WORKDIR /app

COPY package.json .
COPY yarn.lock .
RUN yarn install

COPY . .

RUN yarn build

FROM nginx:1.21.3-alpine

COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80

ENTRYPOINT ["nginx"]

CMD ["-g", "daemon off;"]
