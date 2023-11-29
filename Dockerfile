FROM gcr.io/distroless/static-debian12:nonroot AS prod
COPY pinny /usr/bin/pinny
ENTRYPOINT ["/usr/bin/pinny"]
LABEL org.opencontainers.image.source https://github.com/koalalab-inc/pinny