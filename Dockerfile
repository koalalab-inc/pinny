FROM gcr.io/distroless/static-debian12:nonroot AS prod
COPY pinny /usr/bin/pinny
ENTRYPOINT ["/usr/bin/pinny"]