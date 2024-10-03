FROM golang:1.22 as build

WORKDIR /cmd

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app -v ./cmd/app

FROM scratch as final
WORKDIR /
COPY --from=build /bin/app /app
EXPOSE 7001
EXPOSE 7002

ENTRYPOINT ["/app"]