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

FROM vanilla AS gcloud
RUN apk add --no-cache \
        python3 \
        curl \
    && curl -L -o /tmp/google-cloud-sdk.tar.gz https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz \
    && mkdir -p /usr/local/share \
    && tar -C /usr/local/share -xvf /tmp/google-cloud-sdk.tar.gz \
    && /usr/local/share/google-cloud-sdk/install.sh \
    && rm /tmp/google-cloud-sdk.tar.gz \
    && rm -rf /var/cache/apk/*
ENV PATH $PATH:/usr/local/gcloud/google-cloud-sdk/bin
ENV GOOGLE_APPLICATION_CREDENTIALS /service-account-key.json
RUN if [ -f /service-account-key.json ]; then rm /service-account-key.json; fi
