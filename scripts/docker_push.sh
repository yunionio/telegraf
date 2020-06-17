#!/bin/bash

set -o errexit
set -o pipefail

readlink_mac() {
  cd `dirname $1`
  TARGET_FILE=`basename $1`

  # Iterate down a (possible) chain of symlinks
  while [ -L "$TARGET_FILE" ]
  do
    TARGET_FILE=`readlink $TARGET_FILE`
    cd `dirname $TARGET_FILE`
    TARGET_FILE=`basename $TARGET_FILE`
  done

  # Compute the canonicalized name by finding the physical path
  # for the directory we're in and appending the target file.
  PHYS_DIR=`pwd -P`
  REAL_PATH=$PHYS_DIR/$TARGET_FILE
}

pushd $(cd "$(dirname "$0")"; pwd) > /dev/null
readlink_mac $(basename "$0")
cd "$(dirname "$REAL_PATH")"
CUR_DIR=$(pwd)
SRC_DIR=$(cd .. && pwd)
popd > /dev/null

DOCKER_DIR="$SRC_DIR/scripts"

REGISTRY=${REGISTRY:-docker.io/yunion}
TAG=${TAG:-latest}

build_bin() {
    docker run --rm \
        -v $GOPATH/src:/root/go/src \
        registry.cn-beijing.aliyuncs.com/yunionio/alpine-build:1.0-1 \
        /bin/sh -c "set -ex; cd /root/go/src/github.com/influxdata/telegraf; make telegraf; chown $(id -u):$(id -g) telegraf"
}

build_image() {
    local tag=$1
    local file=$2
    local path=$3
    sudo docker build -t "$tag" -f "$2" "$3"
}

push_image() {
    local tag=$1
    sudo docker push "$tag"
}

COMPONENTS="telegraf"
TAG="release-1.5"

cd $SRC_DIR
build_bin $COMPONENTS
img_name="$REGISTRY/$COMPONENTS:$TAG"
build_image $img_name $DOCKER_DIR/Dockerfile.$COMPONENTS $SRC_DIR
push_image "$img_name"
