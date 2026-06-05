FROM ubuntu:latest
LABEL authors="Omid"

ENTRYPOINT ["top", "-b"]