FROM index.tenxcloud.com/docker_library/alpine
MAINTAINER yarntime@163.com

ADD build/bin/aiops /usr/local/bin/aiops

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/aiops"]
