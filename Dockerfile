FROM golang:alpine3.13 AS build
WORKDIR /src
COPY . .
RUN GO111MODULE=on go build -o /k8sviz .

FROM alpine:3.11
RUN apk add --no-cache bash graphviz ttf-linux-libertine

COPY icons /icons
COPY --from=build /k8sviz /

CMD /k8sviz
