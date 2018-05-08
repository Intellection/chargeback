FROM golang:1.9.0-alpine

# Install system dependencies
RUN apk --update add \
      # tools
      bash git curl make \
      # go package management
      glide && \
    rm -rf /var/cache/apk/*

# Configure source path
ARG SOURCE_PATH="/go/src/github.com/Intellection/chargeback"
WORKDIR ${SOURCE_PATH}

# Install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Install dependencies
ADD . ${SOURCE_PATH}/
RUN make dependencies

CMD ["/bin/bash"]
