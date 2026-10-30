# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=docker" -o /out/spillcheck ./cmd/spillcheck

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/spillcheck /spillcheck
USER nonroot:nonroot
EXPOSE 7700
HEALTHCHECK --interval=30s --timeout=3s CMD ["/spillcheck", "doctor"]
ENTRYPOINT ["/spillcheck", "up"]
