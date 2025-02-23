FROM alpine:3.19 as builder

# Install build dependencies
RUN apk add --no-cache \
    curl \
    wget \
    git \
    build-base \
    openssl-dev \
    zlib-dev \
    bzip2-dev \
    readline-dev \
    sqlite-dev \
    libffi-dev \
    go  # Use Alpine's native Go package instead

# Install Go tools
RUN go install github.com/charmbracelet/glow@latest && \
    go install github.com/ipinfo/cli/ipinfo@latest

FROM alpine:3.19

# Copy Go binaries from builder
COPY --from=builder /root/go/bin/* /usr/local/bin/

# Installing essential packages
RUN apk add --no-cache \
    bash \
    bash-completion \
    openssh \
    curl \
    wget \
    git \
    vim \
    neovim \
    tmux \
    ripgrep \
    fd \
    fzf \
    bat \
    tree \
    jq \
    nodejs \
    npm \
    python3 \
    py3-pip \
    aws-cli \
    kubectl \
    docker-cli \
    htop \
    iputils \
    bind-tools \
    netcat-openbsd \
    tzdata \
    ca-certificates

# Install Terraform
RUN wget https://releases.hashicorp.com/terraform/1.7.4/terraform_1.7.4_linux_arm64.zip && \
    unzip terraform_1.7.4_linux_arm64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.7.4_linux_arm64.zip

# Install Google Cloud SDK
RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-458.0.0-linux-arm.tar.gz && \
    tar -xzf google-cloud-cli-458.0.0-linux-arm.tar.gz && \
    ./google-cloud-sdk/install.sh --quiet && \
    mv google-cloud-sdk /usr/local/ && \
    rm google-cloud-cli-458.0.0-linux-arm.tar.gz

# Add Google Cloud SDK to PATH
ENV PATH $PATH:/usr/local/google-cloud-sdk/bin

# Rest of your original Dockerfile...
RUN rm -rf /var/cache/apk/*

# Configure SSH
RUN ssh-keygen -A \
    && mkdir -p /root/.ssh \
    && chmod 700 /root/.ssh \
    && echo "PermitRootLogin yes" >> /etc/ssh/sshd_config \
    && echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config

# Set up Python environment
RUN python3 -m venv /root/adminvenv && \
    . /root/adminvenv/bin/activate && \
    pip install --no-cache-dir wheel

# Create aliases for common tools
RUN echo 'alias bat="batcat"' >> /root/.bashrc && \
    echo 'source /root/adminvenv/bin/activate' >> /root/.bashrc && \
    echo 'export PATH=$PATH:/usr/local/google-cloud-sdk/bin' >> /root/.bashrc

# Set up working directory
WORKDIR /workspace

# Expose SSH port
EXPOSE 22

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ps aux | grep sshd || exit 1

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]