FROM golang:1.15 as build-be
ADD . /quickshare
WORKDIR /quickshare
RUN /quickshare/scripts/build_be_docker.sh

FROM node as build-fe
COPY --from=build-be /quickshare /quickshare
WORKDIR /quickshare
RUN yarn install \
    && yarn --cwd "src/client/web" run build \
    && cp -R /quickshare/public /quickshare/dist/quickshare

FROM debian:stable-slim
COPY --from=build-fe /quickshare/dist/quickshare /quickshare
ADD configs/docker.yml /quickshare
CMD ["/quickshare/start", "-c", "/quickshare/docker.yml"]
