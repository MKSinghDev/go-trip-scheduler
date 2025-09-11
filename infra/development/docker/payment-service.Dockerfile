FROM alpine
WORKDIR /app

COPY shared shared
COPY build build

ENTRYPOINT build/payment-service
