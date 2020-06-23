# Build:
#   docker build \
#     [--build-arg USER_ID=<host_uid>] \
#     [--build-arg GROUP_ID=<host_gid>] \
#     [--build-arg USERNAME=<username>] \
#     -t ebook-convert .
# Run:
#   docker run \
#     -ti --rm \
#     -v $PWD:/ebooks \
#     ebook-convert

FROM debian:buster-slim

ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=calibre

RUN apt-get update
RUN apt-get install -y calibre
RUN apt-get clean

RUN groupadd -g ${GROUP_ID} ${USERNAME} && \
    useradd -u ${USER_ID} -g ${USERNAME} -m ${USERNAME}

USER ${USERNAME}

WORKDIR /ebooks

ENTRYPOINT ["/usr/bin/ebook-convert"]
