# Glime
#
# Run a glime mirror

FROM centos
MAINTAINER Geoffrey Hayes <hayesgm@gmail.com>

RUN yum install git-core -y

RUN git clone https://github.com/hayesgm/glime /srv/glime && cd /srv/glime && git reset --hard 0.0.1

CMD cd /srv/glime && ./glime.linux

EXPOSE 80:1111