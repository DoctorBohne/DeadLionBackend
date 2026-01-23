FROM ubuntu:latest
LABEL authors="benito"

ENTRYPOINT ["top", "-b"]