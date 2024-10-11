FROM golang:1.22.5 AS builder

WORKDIR /app/

COPY . .

RUN go mod download \
    && CGO_ENABLED=0 go build -o ./main ./main.go

#-------------------------------------------------------------------#
FROM golang:1.22.5

WORKDIR /app/

RUN useradd somebody -u 1001 --create-home --user-group \
  && apt-get update \
  && apt-get install -y pkg-config \
  && rm -rf /var/lib/apt/lists/* \
  && echo "Asia/Bangkok" > /etc/timezone \
  && ln -snf /usr/share/zoneinfo/Asia/Bangkok /etc/localtime \
  && dpkg-reconfigure -f noninteractive tzdata

COPY --from=builder /app/main .

CMD ["./main"]
