FROM golang:1.22-bookworm AS builder

ENV STIFF_VERSION 1.1.0

WORKDIR /go/src
RUN wget -O stiff-${STIFF_VERSION}.tar.gz https://github.com/lezgomatt/stiff/archive/v${STIFF_VERSION}.tar.gz \
  && tar xf stiff-${STIFF_VERSION}.tar.gz \
  && mv stiff-${STIFF_VERSION} stiff

WORKDIR /go/src/stiff
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s"

################################
FROM debian:bookworm-slim

COPY --from=builder /go/src/stiff/stiff /usr/bin/stiff

EXPOSE 1717

CMD ["stiff"]
