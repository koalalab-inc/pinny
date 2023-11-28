# Pinned golang:alpine using pinny
FROM golang@sha256:110b07af87238fbdc5f1df52b00927cf58ce3de358eeeb1854f10a8b5e5e1411 AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp

# Pinned alpine:latest using pinny on Tue, 28 Nov 2023 13:49:54 IST
FROM alpine@sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978 
WORKDIR /app
COPY --from=builder /app/myapp .
EXPOSE 8080
CMD ["./myapp"]
