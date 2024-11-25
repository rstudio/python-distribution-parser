FROM ubuntu:noble
ARG DEBIAN_FRONTEND=noninteractive
ARG OS_IDENTIFIER=ubuntu-2404

# Install necessary packages
RUN apt-get update -qq && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    gnupg \
    gnupg1 \
    git && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

ENV PYTHON_VERSION="3.12.4"
ENV PATH=/opt/python/"${PYTHON_VERSION}"/bin:$PATH
RUN ARCH=$(dpkg --print-architecture) && \
    case $ARCH in \
        amd64) \
            curl -O https://cdn.rstudio.com/python/ubuntu-2404/pkgs/python-${PYTHON_VERSION}_1_amd64.deb && \
            apt-get update -qq && \
            apt-get install -f -y ./python-${PYTHON_VERSION}_1_amd64.deb && \
            rm python-${PYTHON_VERSION}_1_amd64.deb && \
            ln -s /opt/python/${PYTHON_VERSION}/bin/pip /usr/local/bin/pip \
            ;; \
        arm64) \
            apt-get update -qq && \
            apt-get install -y python3 \
            python3-software-properties \
            python3-unittest2 \
            python3-build \
            python3-virtualenv \
            python3-venv && rm -rf /usr/lib/python3.12/EXTERNALLY-MANAGED && \
            # Install `pip` ignoring PEP 668
            curl -sS https://bootstrap.pypa.io/get-pip.py -o get-pip.py && \
            python3 get-pip.py && \
            rm get-pip.py \
            ;; \
        *) echo "unsupported architecture" >&2; exit 1 ;; \
    esac && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

# Install Python dependencies
RUN pip install --upgrade pip && \
    pip install \
        pipenv==v2024.0.1 \
        build \
        virtualenv \
        wheel \
        twine && \
    ln -s /opt/python/${PYTHON_VERSION}/bin/twine /usr/local/bin/twine || echo 'twine symlink failed'

# Install Go with checksum verification and dependencies
ARG GOBIN=/usr/local/bin
ENV PATH="$PATH:/usr/bin:/usr/local/go/bin" \
    GOLANG_VERSION=1.23.2 \
    GOLANG_SHA256_X86=542d3c1705f1c6a1c5a80d5dc62e2e45171af291e755d591c5e6531ef63b454e \
    GOLANG_SHA256_ARM=f626cdd92fc21a88b31c1251f419c17782933a42903db87a174ce74eeecc66a9

RUN ARCH=$(dpkg --print-architecture); \
    case "$ARCH" in \
        amd64) GOLANG_ARCH="amd64"; SHA256="$GOLANG_SHA256_X86" ;; \
        arm64) GOLANG_ARCH="arm64"; SHA256="$GOLANG_SHA256_ARM" ;; \
        *) echo >&2 "unsupported architecture: $ARCH"; exit 1 ;; \
    esac; \
    curl -fsSL "https://dl.google.com/go/go${GOLANG_VERSION}.linux-${GOLANG_ARCH}.tar.gz" -o golang.tar.gz; \
    echo "${SHA256} *golang.tar.gz" | sha256sum -c -; \
    tar -C /usr/local -xzf golang.tar.gz; \
    rm golang.tar.gz && \
    go install gotest.tools/gotestsum@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.1 && \
    go clean -cache


# Install Rust
RUN curl https://sh.rustup.rs -sSf | \
    sh -s -- --default-toolchain stable -y


RUN mkdir /python-distribution-parser
# Start in `/python-distribution-parser` which should be mounted in
WORKDIR /python-distribution-parser
