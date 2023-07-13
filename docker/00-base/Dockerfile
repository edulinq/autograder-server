FROM ubuntu:22.04@sha256:83f0c2a8d6f266d687d55b5cb1cb2201148eb7ac449e4202d9646b9083f1cee0

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update

RUN apt-get -y install \
    locales \
    tzdata

RUN echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
RUN echo "LANG=en_US.UTF-8" > /etc/locale.conf
RUN locale-gen

RUN apt-get install -y \
    curl \
    git \
    python3-pip \
    python3-setuptools \
    time \
    tree \
    vim \
    wget

RUN mkdir -p /autograder/build
RUN mkdir -p /autograder/work
RUN mkdir -p /autograder/submission
RUN mkdir -p /autograder/output

WORKDIR /autograder/work
