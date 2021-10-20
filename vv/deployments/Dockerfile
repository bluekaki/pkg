FROM golang:1.17.2-alpine3.14 AS builder

ADD https://github.com/protocolbuffers/protobuf/releases/download/v3.18.1/protoc-3.18.1-linux-x86_64.zip /opt
ADD https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.6.0/protoc-gen-grpc-gateway-v2.6.0-linux-x86_64 /go/bin/protoc-gen-grpc-gateway
ADD https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.6.0/protoc-gen-openapiv2-v2.6.0-linux-x86_64 /go/bin/protoc-gen-openapiv2
ADD https://raw.githubusercontent.com/bluekaki/pkg/main/vv/pkg/plugin/interceptor/options/options.proto /opt/options.proto
ADD https://raw.githubusercontent.com/bluekaki/pkg/main/vv/deployments/entrypoint.sh /opt/entrypoint.sh

RUN apk update && apk add --no-cache git
RUN cd /opt && unzip protoc-3.18.1-linux-x86_64.zip
RUN cd /opt && git clone https://github.com/googleapis/googleapis.git && \
    cd googleapis && git checkout -b 69697504d9eba1d064820c

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install github.com/bluekaki/pkg/vv/pkg/plugin/protoc-gen-message-validator@latest
RUN chmod +x /go/bin/protoc-gen-grpc-gateway /go/bin/protoc-gen-openapiv2
RUN chmod +x /opt/entrypoint.sh


FROM busybox:glibc

COPY --from=builder /go/bin /usr/local/bin
COPY --from=builder /opt/bin /usr/local/bin
COPY --from=builder /opt/include /usr/local/include
COPY --from=builder /opt/googleapis/google/api /usr/local/include/google/api
COPY --from=builder /opt/options.proto /usr/local/include/options.proto
COPY --from=builder /opt/entrypoint.sh /entrypoint.sh

VOLUME ["/opt"]
WORKDIR /opt

ENTRYPOINT ["/entrypoint.sh"]
