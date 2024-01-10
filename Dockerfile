FROM scratch

ADD ubuntu-focal-core-cloudimg-amd64-root.tar.gz /

WORKDIR /nas_server

RUN apt update -y && \
    apt install -y wget curl git

# ENV GO_VERSION 1.19
# ENV GOROOT /usr/local/go
# ENV GOPATH /go
# ENV PATH ${GOPATH}/bin:/usr/local/go/bin:${PATH}
# RUN wget https://golang.google.cn/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
#     tar -C /usr/local/ -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
#     rm go${GO_VERSION}.linux-amd64.tar.gz

# RUN apt-get install -y make
RUN apt install -y lsb-release curl gpg && \
    curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list
RUN apt update -y && \
    apt install -y redis
RUN apt install -y mysql-server

COPY run.sh /
COPY nas_server /
COPY html /html

RUN mkdir logs

RUN chmod +x nas_server && \
    chmod +x run.sh

EXPOSE 9000
# EXPOSE 6379

CMD sh run.sh