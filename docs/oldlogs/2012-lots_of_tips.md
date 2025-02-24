---
created_at: 2012-07-09
excerpt: my own documentation of lots of tips
kind: article
publish: true
tags: [doc, tip, updating]
title: "lots of tips"
---

### __cmd | 命令行__
* repeat last command argument
        !$ last argument
        !* all argument
* notify cmd line this is the end of options
        --
* readline
        Ctrl-a, Ctrl-e, Ctrl-u, Ctrl-k
        Ctrl-w  delete a word backwards
        Ctrl-x Ctrl-e actually editing the command line
        Esc-f/Esc-b
* get public ip
        curl http://ip4.me 2>/dev/null | sed -e 's#<[^>]*>##g' | grep '^[0-9]'
* http request and response using curl
        curl -iv ${url}
* lock tty
        clear && vlock -c
* real path
        pwd -P
* grep and highlight finds in full file
        grep --color -E "test|$" yourfile
* awk check characters at certain position
        awk -F '' '$25 == "a"' input >output
* iptables
        sudo iptables -I INPUT -p tcp --dport 25 -j ACCEPT
        sudo iptables-save
* mysqld
        repair all the databases
        mysqlcheck -Aor




### __text process | 文本處理__
* print a line in file
        $ sed -n 4p $file_name
        -n switch turns off default output
        $ sed -n 3,60p $file_name
        $ sed -n -e 3,60p -e 1p $file_name


### __vim | basic vim__
* directory and files
        - open directory listing in vim
        :e.
        - open file in this directory listing
        o
* fix sudo
        :w !sudo tee %
* read outline from cmd
        :r!date
* read content from another file
        :read other_file
        :r other_file
* edit and read the template
        :e newfile.cpp | r template.cpp


### __app | 應用__
* chrome
        - keybinding to next tab in chrome: Ctrl-Tab
        - keybinding to next window in chrome: Cmd-`
* curl
        - debug it: curl -vv https://s3.amazonaws.com/

* tmux
        - [un]synchronize: setw synchronize-panes
* tty

* mac osx
        - screenshot screen capture ⌘ ⇧ 3
        - networksetup
                -listallnetworkservices # 列表所有的interface/setup
                # 使用pacfile # need to disable "send all traffic over VPN connection" in VPN setup


* combine two images (merge)
        $> convert +append a.png b.png c.png
        $> montage -geometry 100% left.jpg right.jpg merged.jpg



### __git | just git__
* git checkout and track remote branch | git 檢出遠程的分支並track
        % git checkout -t origin/remote_feature
* git remove deleted files | git rm 刪除物理刪除的文件
        % git status |grep deleted |awk '{print $3}' | xargs git rm
* git setup upstream default (to origin/master) | git 設置缺省上行分支到 origin/master
        % git config branch.master.remote origin
        % git config branch.master.merge refs/heads/master
* git push to current tracking branch (instead of default all branches) | git push 推到目前的trakcing分支(而不是所有分支)
        % git config --global push.default tracking
* git find mod between local and remote | git 檢查本地和遠程mod
        % git fetch
        % git log ..@{u} --stat
        % git log @{u}..


### __tools | 工具箱__
* centos/yum
* brew
* python
        fabric
        ipython
        gunicorn
* pip
        curl -O http://python-distribute.org/distribute_setup.py
        python distribute_setup.py
        easy_install pip

* virtualization
        chef/puppet/vagrant/cucumber/aws/ec2/boto
* rvm/bundle

### __ideas | 思路__
* http stack
        http request > web server > router > data handling > response
* python webapp stack
        django + flask + nginx +uWSGI


### __spartan life hacks__ (which i dont want to forget)
* mental
        mental dicipline
        modest and sincere
        dont judge 不下結論
* physical
        balance/cadence/flex
        @sleeping crunch/lay-biking/ab round/ab breathe
        @standing backkick/ab breathe 腹式呼吸
        @running correct overstriding
        @eating water/vegie





### __links | 總是有用的連接__
* [unix toolbox][link_toolbox]
* [ex:coloer picking][link_colorlover]
* [12 factor for devop][link_12fac]
* [pep][link_pep]
* [rfc][link_rfc] or [ietf version of rfc][link_rfc2]
* [my own devop requirements.txt][link_pipbasic]


---

__go [home][link_home] lets go back back to [home][link_home]__
[link_colorlover]: http://www.colourlovers.com/ "Title: colorlover"
[link_12fac]: http://www.12factor.net/ "Title: 12 factor"
[link_pep]: http://www.python.org/dev/peps/ "Title: python PEPs"
[link_rfc]: http://www.faqs.org/rfcs/ "Title: RFC"
[link_rfc2]: http://tools.ietf.org/html/ "Title: RFC2"



