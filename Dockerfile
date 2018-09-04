FROM golang:1.11 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN update-ca-certificates

WORKDIR $GOPATH/src/github.com/travis-ci/macbot
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM scratch
COPY --from=builder /app .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./app"]
