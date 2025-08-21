FROM --platform=$BUILDPLATFORM golang:1.24@sha256:a18e9e0d94dcc4fffb5c6fa5ec8580edd2e8adcf6541e53304f2bfcaafafd52e

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build/
COPY main.go go.mod go.sum ./
COPY http/ ./http
COPY html/ ./html
COPY app/ ./app

ENV CGO_ENABLED=0
RUN go mod download
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o consensus

FROM scratch

EXPOSE 8088
WORKDIR /app

COPY --from=0 /build/consensus ./

ENTRYPOINT [ "./consensus" ]
