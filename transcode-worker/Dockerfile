FROM  ubuntu:16.04

ENV  DEBIAN_FRONTEND noninteractive
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV GOOS linux

COPY . /go/src/transcode/transcode-worker

WORKDIR /go/src/transcode/transcode-worker

ADD https://github.com/openfaas/faas/releases/download/0.7.9/fwatchdog /usr/bin 

RUN  chmod +x /usr/bin/fwatchdog && \
    apt-get update && \
    apt-get -y install \
      software-properties-common wget git  && \
    add-apt-repository ppa:stebbins/handbrake-releases && \
    apt-get update && \
    apt-get -y install \
      ffmpeg \
      frei0r-plugins \
      handbrake-cli \
      mediainfo \
      mencoder \
      mkvtoolnix \
      mp4v2-utils \
      mplayer2 \
      ruby \
      sshfs \
    && \
    gem install video_transcoding && \
    wget https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz && \
    tar -xvf go1.10.1.linux-amd64.tar.gz && \
    mv go /usr/local && \
    mv /usr/local/go/bin/go /usr/local/bin && \
    go get && \
    go build -o /usr/local/bin/transcode-worker && \
    apt-get clean && \
    rm -rf /var/cache/* /var/log/apt/* /tmp/* && \
    rm go1.10.1.linux-amd64.tar.gz

ENV fprocess "/usr/local/bin/transcode-worker"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["/usr/bin/fwatchdog"]

ENV write_debug="true"
