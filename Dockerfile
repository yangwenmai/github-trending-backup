FROM golang:1.10.2

WORKDIR /go/src/app
COPY . .

RUN mkdir -p /root/.ssh

COPY ssh_config/config /root/.ssh/config
COPY ssh_config/known_hosts /root/.ssh/known_hosts
COPY ssh_config/id_rsa_yangwenmai /root/.ssh/id_rsa_yangwenmai
COPY ssh_config/id_rsa_yangwenmai.pub /root/.ssh/id_rsa_yangwenmai.pub

RUN go install -v ./...

CMD [ "app" ]
