FROM golang:1.25-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /rtmp .

FROM alpine:3

ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /rtmp .
RUN mkdir runtime

EXPOSE 1935 8080

CMD ["./rtmp", "-rtmp", "0.0.0.0:1935", "-http", "0.0.0.0:8080"]
