# ~/.bashrc - Interactive shell configuration
# This handles all interactive shell settings

# If not running interactively, don't do anything
case $- in
    *i*) ;;
      *) return;;
esac

# History settings
HISTCONTROL=ignoreboth
HISTSIZE=10000
HISTFILESIZE=20000
HISTTIMEFORMAT='%Y-%m-%d %H:%M.%S | '
HISTIGNORE="ls:exit:history:[bf]g:jobs"
shopt -s histappend

# Basic aliases for all environments
alias vi=nvim
alias cp="cp -i"
alias cat=batcat
alias bat=batcat
alias less="less -r"
alias ls="ls --color=auto"
alias ll="ls --color"
alias weather='curl wttr.in'
alias vpn='curl ifconfig.me;echo'
alias config='git --git-dir=$HOME/.mydot/ --work-tree=$HOME'

# Development environment
[ -f "$HOME/.cargo/env" ] && source "$HOME/.cargo/env"
[ -f "$HOME/.deno/env" ] && source "$HOME/.deno/env"
[ -f "$HOME/.env" ] && source "$HOME/.env"

# Source modular configurations
for config in ~/.bash/{aws,go,history,prompt,work_alias}.bash; do
    [ -f "$config" ] && source "$config"
done

# OS-specific configurations
if [[ "$OSTYPE" == "darwin"* ]]; then
    [ -f ~/.bash/macos.bash ] && source ~/.bash/macos.bash
else
    # Linux Completion
    [ -f "/usr/share/bash-completion/bash_completion" ] && . "/usr/share/bash-completion/bash_completion"
    [ -f "/usr/local/bin/eksctl" ] && source <(eksctl completion bash)
    [ -f "$HOME/.local/share/bash-completion/completions/deno.bash" ] && source "$HOME/.local/share/bash-completion/completions/deno.bash"
fi

# Tool-specific completions (common to both OS)
command -v github-copilot-cli >/dev/null && eval "$(github-copilot-cli alias -- "$0")"

# Virtual environment (if exists)
[ -f "${HOME}/adminvenv/bin/activate" ] && source "${HOME}/adminvenv/bin/activate"
