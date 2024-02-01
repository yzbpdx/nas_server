FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

WORKDIR /server

RUN apt update -y && \
    apt install -y wget curl git

RUN apt install -y lsb-release curl gpg && \
    curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list
RUN apt update -y && \
    apt install -y redis
RUN apt install -y mysql-server

COPY run.sh /server
COPY script.sql /server
COPY nas_server /server
COPY html /server/html
COPY conf/server.yaml /server/conf/server.yaml

RUN mkdir logs

RUN chmod +x /server/nas_server && \
    chmod +x /server/run.sh

EXPOSE 9000

CMD sh /server/run.sh
