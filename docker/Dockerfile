FROM golang:1.14-buster AS builder

ENV STIFF_VERSION 1.0.1

WORKDIR /go/src
RUN wget -O stiff-${STIFF_VERSION}.tar.gz https://github.com/undecidabot/stiff/archive/v${STIFF_VERSION}.tar.gz \
  && tar xf stiff-${STIFF_VERSION}.tar.gz \
  && mv stiff-${STIFF_VERSION} stiff

WORKDIR /go/src/stiff
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s"

################################
FROM debian:buster-slim

COPY --from=builder /go/src/stiff/stiff /usr/bin/stiff

EXPOSE 1717

CMD ["stiff"]
