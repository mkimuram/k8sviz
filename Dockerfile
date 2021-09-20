FROM golang:alpine3.13 AS build
RUN apk add --no-cache make
WORKDIR /src
COPY . .
RUN make build

FROM alpine:3.11 AS vanilla
RUN apk add --no-cache bash graphviz ttf-linux-libertine

COPY icons /icons
COPY --from=build /src/bin/k8sviz /

CMD /k8sviz

FROM vanilla AS aws
RUN apk add --no-cache \
        python3 \
        py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install \
        awscli \
    && rm -rf /var/cache/apk/*
