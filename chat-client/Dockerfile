FROM node:20.13-alpine AS builder

WORKDIR /app

COPY ./package.json ./

RUN npm install --verbose

COPY . .

RUN npm run build

FROM nginx:1.26-alpine

WORKDIR /usr/share/nginx/html

EXPOSE 8080

COPY nginx/nginx.conf /etc/nginx/nginx.conf

COPY nginx/conf.d /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html
