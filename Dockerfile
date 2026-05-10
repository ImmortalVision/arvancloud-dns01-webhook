FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/webhook ./main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/webhook /webhook

ENTRYPOINT ["/webhook"]
