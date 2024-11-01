FROM golang

WORKDIR /pb

RUN go env -w GOCACHE=/go-cache

RUN go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/gomod-cache

RUN go mod download

COPY *.go ./

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache

RUN CGO_ENABLED=0 GOOS=linux go build -o /pocketbase

CMD ["/pocketbase", "serve", "--http=0.0.0.0:8090"]