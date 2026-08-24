FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/dhcp-ipam-control ./cmd/server
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/dhcp-ipam-control /dhcp-ipam-control
EXPOSE 8080 6767/udp 6768/udp
ENTRYPOINT ["/dhcp-ipam-control"]
