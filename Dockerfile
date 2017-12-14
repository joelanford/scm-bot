FROM alpine:3.6
MAINTAINER Joe Lanford <joe.lanford@gmail.com>

ARG GIT_COMMIT
LABEL maintainer="Joe Lanford <joe.lanford@gmail.com>" gitCommit="${GIT_COMMIT}"

ADD scm-bot /usr/local/bin/scm-bot
ENTRYPOINT [ "/usr/local/bin/scm-bot" ]
