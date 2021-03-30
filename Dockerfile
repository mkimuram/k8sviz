FROM golang:alpine3.13 AS build
RUN apk add --no-cache make
WORKDIR /src
COPY . .
RUN make build

FROM alpine:3.11
RUN apk add --no-cache bash graphviz ttf-linux-libertine curl python
# Downloading gcloud package
RUN curl https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz > /tmp/google-cloud-sdk.tar.gz

# Installing the package
RUN mkdir -p /usr/local/share \
  && tar -C /usr/local/share -xvf /tmp/google-cloud-sdk.tar.gz \
  && /usr/local/share/google-cloud-sdk/install.sh

# Adding the package path to local
ENV PATH $PATH:/usr/local/gcloud/google-cloud-sdk/bin
ENV GOOGLE_APPLICATION_CREDENTIALS /service-account-key.json

RUN if [ -f /service-account-key.json ]; then rm /service-account-key.json; fi
COPY icons /icons
COPY --from=build /src/bin/k8sviz /

CMD /k8sviz
