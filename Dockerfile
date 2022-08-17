# FROM zenika/alpine-chrome:latest as CHROME

FROM node:alpine as NODE_DEPS
WORKDIR /app

ENV PUPPETEER_SKIP_DOWNLOAD=1
COPY ./package.json ./package.json
COPY ./package-lock.json ./package-lock.json
RUN npm install

FROM golang:alpine as GO_BUILDER
WORKDIR /app
COPY ./go.mod ./go.mod
COPY ./main.go ./main.go
RUN go mod tidy
RUN go build -o main .


# mermaid bin (mmdc) requires `node`...
FROM node:alpine as RUNNER
WORKDIR /app

# install chromium-browser -> /usr/bin/chromium-browser
RUN apk add chromium

# https://github.com/browserless/chrome/blob/master/Dockerfile#L13
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium-browser

# See issue for flags
# https://github.com/adieuadieu/serverless-chrome/issues/170#issuecomment-430485464
COPY puppeteer-config.json ./puppeteer-config.json

COPY --from=NODE_DEPS /app/package.json ./package.json
COPY --from=NODE_DEPS /app/package-lock.json ./package-lock.json
COPY --from=NODE_DEPS /app/node_modules ./node_modules

# Copy our binary to the container
COPY --from=GO_BUILDER /app/main ./main

RUN mkdir -p out
RUN chmod 777 out

ENTRYPOINT [ "/bin/sh", "-l", "-c" ]

# Compatibility step w/ Lambda Function URL or API Gateway
EXPOSE 8080
ENV READINESS_CHECK_PORT=8080
ENV PORT=8080
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.3.3 /lambda-adapter /opt/extensions/lambda-adapter

# Execute our binary
CMD [ "./main" ]
