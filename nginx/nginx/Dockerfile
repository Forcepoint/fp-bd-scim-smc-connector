FROM nginx:1.13.0-alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY conf.d /etc/nginx/conf.d/
RUN apk add --no-cache nginx-mod-http-perl=1.10.3-r1