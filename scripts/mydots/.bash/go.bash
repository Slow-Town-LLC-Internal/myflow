

[ -d /usr/local/go/bin ] && export PATH="/usr/local/go/bin:$PATH" && export GOROOT="/usr/local/go"
[ -d ${HOME}/go ] && export GOPATH="${HOME}/go" && export PATH="$GOPATH/bin:$PATH"
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org


# go install github.com/charmbracelet/glow@latest

