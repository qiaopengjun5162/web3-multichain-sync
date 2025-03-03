#!/usr/bin/env bash

function exit_if() {
  local extcode="$1"
  local msg="$2"
  if [ "$extcode" -ne 0 ]; then
    if [ -n "$msg" ]; then
      echo "$msg" >&2
    fi
    exit "$extcode"
  fi
}

# 检查 protoc-gen-go 和 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Protocol Buffers plugin for Go is not installed. Please install it using:" >&2
    echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" >&2
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "gRPC plugin for Go is not installed. Please install it using:" >&2
    echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" >&2
    exit 1
fi

echo "Compiling Go interfaces..."

# 设置环境变量
GOBIN=$(go env GOBIN)
exit_if $? "Failed to get GOBIN from go env."

export GOBIN
export PATH=$PATH:$GOBIN

# 编译 .proto 文件
protoc -I ./protobuf --go_out=./ --go-grpc_out=require_unimplemented_servers=false:. ./protobuf/*.proto
exit_if $? "Failed to compile .proto files."

echo "Done."
