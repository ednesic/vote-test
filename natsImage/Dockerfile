FROM alpine

# Setting up environment
ENV NATS_VERSION=0.11.0

# Install gnatsd
RUN apk add --update --no-cache ca-certificates \
    && update-ca-certificates \
    && apk add --no-cache openssl \
    && wget https://github.com/nats-io/nats-streaming-server/releases/download/v${NATS_VERSION}/nats-streaming-server-v${NATS_VERSION}-linux-386.zip \
    && unzip nats-streaming-server-v${NATS_VERSION}-linux-386.zip \
    && mv nats-streaming-server-v${NATS_VERSION}-linux-386/nats-streaming-server /usr/bin/nats-streaming-server \
    && rm nats-streaming-server-v${NATS_VERSION}-linux-386.zip \
    && rm -r nats-streaming-server-v${NATS_VERSION}-linux-386 \
    && apk del ca-certificates openssl

# Copy configuration and entrypoint
COPY entrypoint.sh /entrypoint.sh
COPY gnatsd.conf /etc/gnatsd.conf

# Secure configuration and entrypoint
RUN chmod 600 /etc/gnatsd.conf \
    && chmod 500 /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["nats-streaming-server", "-c", "/etc/gnatsd.conf"]
