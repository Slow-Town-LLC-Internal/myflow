# ~/.bash_profile - Login shell initialization
# This should be minimal and primarily handle environment setup

# Source .bashrc for interactive shells
[[ -r ~/.bashrc ]] && . ~/.bashrc

# Environment variables that should be set once
export LANG="en_US.UTF-8"
export TERM="xterm-256color"
export CLICOLOR=1
export PAGER="less -r"

# Path management function (define here as it's needed early)
path_append() {
    if [ -d "$1" ] && [[ ":$PATH:" != *":$1:"* ]]; then
        PATH="${PATH:+"$PATH:"}$1"
    fi
}

# Core PATH setup
path_append "$HOME/bin"
path_append "$HOME/.local/bin"
path_append "$HOME/local/bin"
path_append "$HOME/local/sbin"
