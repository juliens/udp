FROM ubuntu
COPY ./dns /dns
ENTRYPOINT /dns
