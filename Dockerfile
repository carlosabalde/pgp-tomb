FROM golang:1.13

RUN apt-get update -y && \
    apt-get install -y less gpg gpg-agent nano vim xsel

COPY files/keys/alice.pri /root
RUN gpg --batch --yes --passphrase 's3cr3t' --import /root/alice.pri && \
    rm /root/alice.pri

RUN echo 'export GPG_TTY=$(tty)' >> /root/.bashrc

RUN go get golang.org/x/tools/cmd/goimports

CMD tail -f /dev/null
