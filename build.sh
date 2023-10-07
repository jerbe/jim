#!/usr/bin/env bash

APP_NAME=jim
BIN=./bin
TARGETS=(
  "darwin/arm64" \
  "darwin/amd64" \
  "freebsd/386" \
  "freebsd/amd64" \
  "freebsd/arm" \
  "freebsd/arm64" \
  "linux/386" \
  "linux/amd64" \
  "linux/arm" \
  "linux/arm64" \
  "linux/mips" \
  "linux/mips64" \
  "windows/386" \
  "windows/amd64" \
  "windows/arm" \
  "windows/arm64"
)


function build_func {
  suffix=""
  if [ "$1" = "windows" ] ; then
    suffix=".exe"
  fi
  echo "正在编译...  系统:[$1] 指令集:[$2]"
  cmdline="GOOS=$1 GOARCH=$2 go build -o ${BIN}/${APP_NAME}-$1-$2${suffix} ./*.go"
  echo "${cmdline}"
  eval ${cmdline}
}

if [ $# -eq 0 ]; then
    echo "没有任何参数"
    exit 1
fi

if [ "$1" = "build" ]; then
  if [ "$2" = "" ]; then
    for ((i=0; i<${#TARGETS[@]}; i++)); do
      parts=$(echo "${TARGETS[$i]}" | awk -F'/' '{ for(i=1; i<=NF; i++) print $i }')
      build_func ${parts[0]} ${parts[1]}
    done
  else
    parts=$(echo "$2" | awk -F'/' '{ for(i=1; i<=NF; i++) print $i }')
    build_func ${parts[0]} ${parts[1]}
  fi
fi


if [ "$1" = "clean" ]; then
  cmdline="rm ${BIN}/${APP_NAME}-*"
  echo $cmdline
  eval $cmdline
fi
