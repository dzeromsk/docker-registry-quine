ARG GO_VERSION=1.20

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o quine

FROM scratch
COPY --from=builder src/quine /quine
ENTRYPOINT ["/quine"]
