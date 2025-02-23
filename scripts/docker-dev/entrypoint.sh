#!/bin/sh
set -e

# Function to load configs
load_configs() {
    CONFIG_DIR="/etc/dev-config"
    
    # Load bash configurations
    if [ -f "$CONFIG_DIR/.bashrc" ]; then
        cp "$CONFIG_DIR/.bashrc" /root/.bashrc
    fi
    
    # Load vim configurations
    if [ -f "$CONFIG_DIR/.vimrc" ]; then
        cp "$CONFIG_DIR/.vimrc" /root/.vimrc
    fi
    
    # Load tmux configurations
    if [ -f "$CONFIG_DIR/.tmux.conf" ]; then
        cp "$CONFIG_DIR/.tmux.conf" /root/.tmux.conf
    fi
    
    # Load git configurations
    if [ -f "$CONFIG_DIR/.gitconfig" ]; then
        cp "$CONFIG_DIR/.gitconfig" /root/.gitconfig
    fi
    
    # Load any custom scripts
    if [ -d "$CONFIG_DIR/scripts" ]; then
        cp -r "$CONFIG_DIR/scripts" /root/
        chmod +x /root/scripts/*
    fi
    
    # Load SSH configurations if present
    if [ -d "$CONFIG_DIR/ssh" ]; then
        cp -r "$CONFIG_DIR/ssh/"* /root/.ssh/
        chmod 600 /root/.ssh/*
    fi
}

# Load all configs
load_configs

# Start SSH daemon
exec /usr/sbin/sshd -D