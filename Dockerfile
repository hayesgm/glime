# Glime
#
# Run a glime mirror

FROM centos
MAINTAINER Geoffrey Hayes <hayesgm@gmail.com>

RUN yum install git-core -y

RUN git clone https://github.com/hayesgm/glime /srv

CMD cd /srv/glime && ./glime.linux

EXPOSE 80