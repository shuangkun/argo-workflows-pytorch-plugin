FROM golang:1.23 AS builder

WORKDIR /go/src/github.com/shuangkun/argo-workflows-pytorch-plugin
COPY . /go/src/github.com/shuangkun/argo-workflows-pytorch-plugin
ENV GO111MODULE=off
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o argo-pytorch-plugin main.go

FROM alpine:3.10
COPY --from=builder /go/src/github.com/shuangkun/argo-workflows-pytorch-plugin/argo-pytorch-plugin /usr/bin/argo-pytorch-plugin
RUN chmod +x /usr/bin/argo-pytorch-plugin
CMD ["argo-pytorch-plugin"]

