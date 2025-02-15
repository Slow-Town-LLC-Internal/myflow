#!/bin/bash

set -e  # Exit on error

# Detect OS and set variables
setup_env() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Detected MacOS environment"
        PACKAGE_MANAGER="brew"
        PYTHON_VENV_PATH="$HOME/adminvenv"
        INSTALL_CMD="brew install"
        INSTALL_CASK_CMD="brew install --cask"
    else
        echo "Detected Linux environment"
        PACKAGE_MANAGER="apt"
        PYTHON_VENV_PATH="/home/vagrant/adminvenv"
        INSTALL_CMD="sudo apt-get install -y"

        # Add repositories for Linux
        echo "Adding repositories..."
        echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
        echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

        # Add GPG keys
        wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
        curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

        # Update package list
        sudo apt-get update
    fi
}

# Install basic development tools
install_basic_tools() {
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        $INSTALL_CMD bash-completion@2 fzf ripgrep fd bat neovim tmux git gh tree dict jq
        $INSTALL_CMD terraform google-cloud-cli awscli kubernetes-cli podman
        $INSTALL_CASK_CMD visual-studio-code google-chrome virtualbox vagrant
    else
        $INSTALL_CMD neovim tmux git curl ripgrep fd-find tree dict jq bat fzf imagemagick
        $INSTALL_CMD build-essential libssl-dev zlib1g-dev libbz2-dev libreadline-dev
        $INSTALL_CMD libsqlite3-dev libncursesw5-dev xz-utils tk-dev libxml2-dev
        $INSTALL_CMD libxmlsec1-dev libffi-dev liblzma-dev
        $INSTALL_CMD google-cloud-cli terraform podman buildah skopeo
        $INSTALL_CMD bash-completion
    fi
}

# Install and configure language environments
setup_languages() {
    # Python with pyenv
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        $INSTALL_CMD pyenv
    else
        curl https://pyenv.run | bash
    fi

    export PYENV_ROOT="$HOME/.pyenv"
    command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"
    eval "$(pyenv init -)"

    pyenv install 3.12.1 || echo "Python 3.12.1 installation failed"
    pyenv global 3.12.1

    # Go
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        $INSTALL_CMD go
    else
        GO_VERSION="1.22.0"
        wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
        rm -f "go${GO_VERSION}.linux-amd64.tar.gz"
    fi

    go install github.com/charmbracelet/glow@latest
    go install github.com/adamryman/tcpterm@latest
    go install golang.org/x/tools/gopls@latest
    go install github.com/ipinfo/cli/ipinfo@latest

    # Node.js
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        $INSTALL_CMD nvm
    else
        curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
    fi

    mkdir -p ~/.nvm
    export NVM_DIR="$HOME/.nvm"
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        [ -s "/opt/homebrew/opt/nvm/nvm.sh" ] && \. "/opt/homebrew/opt/nvm/nvm.sh"
    else
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    fi

    nvm install 20
    nvm use 20
    npm install -g typescript ts-node
}

# Set up dotfiles
setup_dotfiles() {
    echo "Setting up dotfiles..."
    # Clone repository
    mkdir -p ~/src
    cd ~/src
    if [ ! -d "myflow" ]; then
        git clone git@github.com:Slow-Town-LLC-Internal/myflow.git
    fi
    cd myflow/scripts/mydots


    # Backup existing files
    for file in $(find . -maxdepth 1 -name ".*" -type f); do
        [ -f ~/$file ] && mv ~/$file ~/${file}.backup.$(date +%Y%m%d_%H%M%S)
    done

    # Backup directories
    for dir in .bash .config .vim .w3m; do
        [ -d ~/$dir ] && mv ~/$dir ~/${dir}.backup.$(date +%Y%m%d_%H%M%S)
    done

    # Copy dotfiles
    cp -r .bash .config .vim .w3m ~/.
    cp .bash_profile .bashrc .gitconfig .tmux.conf .vimrc ~/.
}

# Set up Python virtual environment
setup_python_venv() {
    echo "Setting up Python virtual environment at $PYTHON_VENV_PATH"
    python3 -m venv "$PYTHON_VENV_PATH"
    source "$PYTHON_VENV_PATH/bin/activate"
    pip install --upgrade pip wheel
}

setup_db() {
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
      # no db on host - pass
      echo "No database setup on host"
    else
      sudo apt-get install -y postgresql postgresql-contrib

      # Configure PostgreSQL to listen on all interfaces
      sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/14/main/postgresql.conf
      # Allow connections from host machine
      echo "host    all             all             10.0.2.2/32            scram-sha-256" | sudo tee -a /etc/postgresql/14/main/pg_hba.conf
      echo "host    all             all             192.168.56.0/24        scram-sha-256" | sudo tee -a /etc/postgresql/14/main/pg_hba.conf

      # Restart PostgreSQL
      sudo systemctl restart postgresql
    fi
}

# OS-specific configurations
os_specific_config() {
    if [[ "$PACKAGE_MANAGER" == "brew" ]]; then
        # MacOS specific
        defaults -currentHost write -g com.apple.keyboard.modifiermapping.1452-566-0 -array-add '
        <dict>
            <key>HIDKeyboardModifierMappingSrc</key>
            <integer>0x700000039</integer>
            <key>HIDKeyboardModifierMappingDst</key>
            <integer>0x7000000E0</integer>
        </dict>
        '

        # Install and configure Tailscale
        $INSTALL_CMD tailscale
        sudo brew services start tailscale
        sudo tailscale up --exit-node=slowtown-me --exit-node-allow-lan-access=true --auto-update
    else
        # Linux specific
        mkdir -p ~/.local/bin

        echo 'XKBOPTIONS="ctrl:nocaps"' | sudo tee /etc/default/keyboard
        sudo dpkg-reconfigure -f noninteractive keyboard-configuration

        echo 'kernel.unprivileged_userns_clone=1' | sudo tee /etc/sysctl.d/00-local-userns.conf
        sudo systemctl restart procps

        # Create bat symlink
        ln -sf /usr/bin/batcat ~/.local/bin/bat
    fi
}

# Main installation process
main() {
    echo "Starting installation..."
    setup_env
    install_basic_tools
    setup_languages
    setup_dotfiles
    setup_python_venv
    setup_db
    os_specific_config

    echo "Installation complete! Please:"
    echo "1. Log out and log back in for all changes to take effect"
    echo "2. Source your new bash configuration: source ~/.bash_profile"
}

main "$@"
