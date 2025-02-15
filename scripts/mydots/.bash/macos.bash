# MacOS specific configurations

# Homebrew path
eval "$(/opt/homebrew/bin/brew shellenv)"

# Python (pyenv)
export PYENV_ROOT="$HOME/.pyenv"
command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"

# Node Version Manager (nvm)
export NVM_DIR="$HOME/.nvm"
[ -s "/opt/homebrew/opt/nvm/nvm.sh" ] && \. "/opt/homebrew/opt/nvm/nvm.sh"
[ -s "/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm" ] && \. "/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm"

# Go
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Vagrant aliases
alias dev='vagrant ssh'
alias vim='vagrant ssh -c "nvim"'
alias w3m='vagrant ssh -c "w3m"'
alias v='vagrant ssh'  # quick SSH
alias vcode='code --remote ssh-remote+vagrant@localhost /projects'  # VSCode in VM

# Project shortcuts
alias proj='cd ~/Projects'

# MacOS specific aliases
alias showfiles='defaults write com.apple.finder AppleShowAllFiles YES; killall Finder'
alias hidefiles='defaults write com.apple.finder AppleShowAllFiles NO; killall Finder'


# completion for macos
[ -f "$(brew --prefix)/etc/profile.d/bash_completion.sh" ] && source "$(brew --prefix)/etc/profile.d/bash_completion.sh"
[ -f "$(brew --prefix)/etc/profile.d/autojump.sh" ] && source "$(brew --prefix)/etc/profile.d/autojump.sh"
[ -f "$(brew --prefix)/bin/vault" ] && complete -C "$(brew --prefix)/bin/vault" vault
# FZF from Homebrew
[ -f "$(brew --prefix)/opt/fzf/shell/completion.bash" ] && source "$(brew --prefix)/opt/fzf/shell/completion.bash"
[ -f "$(brew --prefix)/opt/fzf/shell/key-bindings.bash" ] && source "$(brew --prefix)/opt/fzf/shell/key-bindings.bash"
