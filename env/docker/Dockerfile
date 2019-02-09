FROM golang:alpine

RUN apk update && apk add --no-cache git curl bash xz python2

RUN adduser nextload --disabled-password --gecos "First Last,RoomNumber,WorkPhone,HomePhone" --home /home/nextload-core

RUN mkdir -p /home/nextload-core/youtube-dl

RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /home/nextload-core/youtube-dl/youtube-dl
RUN chmod a+rx /home/nextload-core/youtube-dl/youtube-dl

RUN curl -L https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz -o /tmp/ffmpeg-release-amd64-static.tar.xz \
  && cd /tmp/ && tar -xJf ffmpeg-release-amd64-static.tar.xz && cp -r /tmp/ffmpeg-4.1-64bit-static/* /home/nextload-core/youtube-dl/ \
  && rm -rf /tmp/ffmpeg-4.1-64bit-static

RUN curl -L https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz -o /tmp/go1.11.5.linux-amd64.tar.gz \
	&& cd /tmp/ \
	&& tar xvf go1.11.5.linux-amd64.tar.gz \
	&& cp /tmp/go/bin/go /usr/local/bin/

ENV PATH $PATH:/usr/local/bin/

RUN /bin/su - nextload -c "mkdir -p /home/nextload-core/go"
ENV GOPATH="/home/nextload-core/go"
RUN /bin/su - nextload -c "mkdir -p /home/nextload-core/bin"
ENV GOBIN="/home/nextload-core/bin"

RUN mkdir -p /home/nextload-core
COPY src/ /home/nextload-core/src
COPY config /home/nextload-core/config
COPY env/docker/scripts /home/nextload-core/scripts

RUN mkdir -p /var/log/nextload-core
RUN mkdir -p /home/nextload-core/bin

RUN chown -R nextload:nextload /home/nextload-core/ && chmod -R u+rwx /home/nextload-core/src
RUN /bin/su - nextload -c "cd /home/nextload-core/src/ && chmod -R u+rwx /home/nextload-core/src && export GOPATH=/home/nextload-core/go && export GOBIN=/home/nextload-core/bin && /usr/local/go/bin/go get -d && /usr/local/go/bin/go install main.go"

RUN chmod u+rx /home/nextload-core/config/config.default.yml 
RUN dos2unix /home/nextload-core/scripts/startup.sh

ENV PATH $PATH:/home/nextload-core/youtube-dl/

ENTRYPOINT ["/bin/bash","/home/nextload-core/scripts/startup.sh"]