# Pinned gcr.io/distroless/static-debian12:nonroot using pinny
FROM gcr.io/distroless/static-debian12@sha256:43a5ce527e9def017827d69bed472fb40f4aaf7fe88c356b23556a21499b1c04 
COPY pinny /usr/bin/pinny
ENTRYPOINT ["/usr/bin/pinny"]
LABEL org.opencontainers.image.source https://github.com/koalalab-inc/pinny
