FROM alpine
MAINTAINER Joe Lanford <joe.lanford@gmail.com>

ADD scm-bot /usr/local/bin/scm-bot
ENTRYPOINT [ "/usr/local/bin/scm-bot" ]
