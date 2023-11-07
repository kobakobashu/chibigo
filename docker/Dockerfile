FROM golang:latest

ARG BRANCH_OR_TAG=main
ARG CMAKE_OPTIONS=
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get install -y \
    make \
    git \
    binutils

ENV WORK_PATH /workspace
WORKDIR $WORK_PATH