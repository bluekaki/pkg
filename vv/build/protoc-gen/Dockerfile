FROM golang:1.15.6-alpine AS builder

ADD https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip /opt
ADD https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.0.1/protoc-gen-grpc-gateway-v2.0.1-linux-x86_64 /go/bin/protoc-gen-grpc-gateway
ADD https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.0.1/protoc-gen-openapiv2-v2.0.1-linux-x86_64 /go/bin/protoc-gen-openapiv2
ADD https://raw.githubusercontent.com/byepichi/pkg/main/vv/options/options.proto /opt/byepichi/pkg/vv/options.proto
ADD https://raw.githubusercontent.com/byepichi/pkg/main/vv/build/protoc-gen/entrypoint.sh /opt/entrypoint.sh

RUN apk update && apk add --no-cache git
RUN cd /opt && unzip protoc-3.14.0-linux-x86_64.zip
RUN cd /opt && git clone https://github.com/googleapis/googleapis.git && \
    cd googleapis && git checkout -b 69697504d9eba1d064820c

RUN go get google.golang.org/protobuf/cmd/protoc-gen-go
RUN go get github.com/byepichi/pkg/protoc-gen-message-validator
RUN go get github.com/byepichi/pkg/vv/cmd/protoc-gen-go-grpc
RUN chmod +x /go/bin/protoc-gen-grpc-gateway /go/bin/protoc-gen-openapiv2
RUN chmod +x /opt/entrypoint.sh


FROM busybox:glibc

COPY --from=builder /go/bin /usr/local/bin
COPY --from=builder /opt/bin /usr/local/bin
COPY --from=builder /opt/include /usr/local/include
COPY --from=builder /opt/googleapis/google/api /usr/local/include/google/api
COPY --from=builder /opt/byepichi/pkg/vv/options.proto /usr/local/include/byepichi/pkg/vv/options.proto
COPY --from=builder /opt/entrypoint.sh /entrypoint.sh

VOLUME ["/opt"]
WORKDIR /opt

ENTRYPOINT ["/entrypoint.sh"]
