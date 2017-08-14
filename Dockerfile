FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/theothertomelliott/pairpad/
RUN go get -d -v golang.org/x/net/html  
COPY .    .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pairpad .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/theothertomelliott/pairpad/ .
ENV PORT 8080
CMD ["./pairpad"]