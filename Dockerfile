FROM golang:1.23.1 as build

WORKDIR /cmd

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app -v ./cmd/app

FROM scratch as final
WORKDIR /
COPY --from=build /bin/app /bin/app
COPY etc/config.env /etc/config.env
COPY etc/dev.env /etc/dev.env

EXPOSE 7001
EXPOSE 8001

ENTRYPOINT ["/bin/app"]