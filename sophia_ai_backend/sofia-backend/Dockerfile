FROM golang:1.24 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /sofia ./cmd/sofia
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /synthesis ./cmd/workers/synthesis

FROM gcr.io/distroless/static-debian12:nonroot AS synthesis
COPY --from=build /synthesis /synthesis
USER nonroot:nonroot
ENTRYPOINT ["/synthesis"]

FROM gcr.io/distroless/static-debian12:nonroot AS sofia
COPY --from=build /sofia /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
