# build stage
FROM golang:1.23.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app .

# run stage
FROM gcr.io/distroless/base-debian12
ENV PORT=8080
COPY --from=build /app /app
COPY --from=build /src/openapi.yaml /openapi.yaml
ENTRYPOINT ["/app"]
