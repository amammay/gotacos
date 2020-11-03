FROM golang:1.15-alpine as builder

WORKDIR /workspace

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

ARG build_target=client

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /app ./$build_target/


# add runsd to the image
ADD https://github.com/ahmetb/runsd/releases/latest/download/runsd /runsd

# gives runsd the correct permissions
RUN chmod +x /runsd


# bare bones distroless image
FROM gcr.io/distroless/base
COPY --from=builder /app /app

# copy runsd to our final image
COPY --from=builder /runsd /runsd

# the most important part is adding runsd before we start our app
ENTRYPOINT ["/runsd", "--", "/app"]
