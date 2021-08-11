FROM golang:1.17 AS build-image

ADD . /src
RUN cd /src && \
    go build -ldflags "-linkmode external -extldflags -static" ./cmd/h2static && \
    strip -s h2static && \
    mkdir /www


FROM scratch

COPY --from=build-image /src/h2static /
COPY --from=build-image /www /www

EXPOSE 8080/tcp
ENTRYPOINT ["/h2static", "-log", "-dir", "/www"]
