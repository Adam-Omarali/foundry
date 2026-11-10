# ---------- Stage 1: build ----------
FROM golang:1.23-alpine AS build
WORKDIR /src

# TODO: copy only the dependency manifests (which two files?)
COPY go.mod go.sum /src/
# TODO: download the modules
RUN go mod download

# TODO: copy the rest of the source code
COPY . /src/

# TODO: build a static Linux binary to /out/api
RUN CGO_ENABLED=0 go build -o /src/out/api
#       (hint: CGO_ENABLED=0 goes in front of `go build`)

# ---------- Stage 2: run ----------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /src/out/api /api
# TODO: copy /out/api from the "build" stage into this image
#       (hint: COPY --from=...)

# TODO: document the port
EXPOSE 8080
# TODO: set the command that starts the binary
ENTRYPOINT ["/api"]
#       (use exec form: ["..."] rather than a plain string)