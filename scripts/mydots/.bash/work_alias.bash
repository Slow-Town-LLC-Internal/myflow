alias ,goeks="python ~/src/myflow/scripts/cluster-switcher.py"
alias ,goaider="cd ~/tmp/scrapt"

### K8s

alias ,getpodver="kubectl get pods -n core -o jsonpath='{range .items[*]}{.metadata.name}{\" : \"}{range .spec.containers[*]}{.image}{\" \"}{end}{\"\n\"}{end}'"
