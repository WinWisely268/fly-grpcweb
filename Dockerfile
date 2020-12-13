# BUILD DOCKER / PODMAN IMAGE
FROM cirrusci/flutter:beta as builder
RUN apt update && \
    add-apt-repository ppa:longsleep/golang-backports
RUN apt install -y ca-certificates build-essential git golang-go libnss3-tools
RUN mkdir -p /usr/local/share/ca-certificates
WORKDIR /go/src/app
COPY . .
RUN make all
#
FROM alpine:latest
WORKDIR /goapp
COPY --from=builder /go/src/app/app /goapp/app
RUN apk --no-cache add ca-certificates
ENV PORT=8080
EXPOSE 8080
CMD ["/goapp/app"]
#LABEL org.opencontainers.image.created="${IMAGE_DATE}" \
#    org.opencontainers.image.title="${IMAGE_NAME}" \
#    org.opencontainers.image.authors="${IMAGE_AUTHOR}" \
#    org.opencontainers.image.revision="${IMAGE_REF}" \
#    org.opencontainers.image.vendor="${IMAGE_ORG}"