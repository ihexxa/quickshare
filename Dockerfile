FROM golang:1.15 as build-be
ADD . /quickshare
WORKDIR /quickshare
RUN /quickshare/scripts/build_exec.sh

FROM node:lts as build-fe
COPY --from=build-be /quickshare /quickshare
WORKDIR /quickshare
RUN yarn run build:fe \
    && cp -R /quickshare/public /quickshare/dist/quickshare

FROM debian:stable-slim
COPY --from=build-fe /quickshare/dist/quickshare /quickshare
ADD configs/docker.yml /quickshare
CMD ["/quickshare/start", "-c", "/quickshare/docker.yml"]
