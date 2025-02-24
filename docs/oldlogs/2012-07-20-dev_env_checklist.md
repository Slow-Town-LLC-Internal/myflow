---
created_at: 2012-07-20
excerpt: my dev env checklist
kind: article
publish: true
tags: [doc, checklist]
title: "開發環境必備checklist"
symbol: ☐ ☑ ☒ 
---

## summary | 簡述

    文檔提供基於linux/macosx環境下開發的個人必備配置
    
## TODO 
    1. finish up the documentation
    2. get together the files linked and fix the links
    3. roll them together use chef or puppet or just shell.
    

## service | 服務

* ☐ dotfiles | 環境
    * ☐ [${HOME}/.gitignore]() [pub]
    * ☐ [${HOME}/.vimrc]() [pub]
    * ☐ [${HOME}/.ssh/auth_keys]() [pub]
    * ☐ [${HOME}/.wgetrc]() [pub]
    * ☐ [${HOME}/.ssh/config]() [pri]
    * ☐ [${HOME}/.ssh/]() [pri]
    * ☐ [${HOME}/.bash_profile]() [pri]
    * ☐ [${HOME}/.ec2]() [pri]
* ☐ servers | 服務器
    * ☐ [linode](http://www.linode.com/)
    * ☐ [aws ec2](http://aws.amazon.com/ec2/)
* ☐ auth | 權限 
    * ☐ [hosts](${HOME}/etc/hosts) [var]
    * ☐ [sshd](${HOME}/etc/ssh/sshd_config) [pri]
    * ☐ [visudo](${HOME}/etc/sudoers) [pri]


## dev | 開發環境

* ☐ pip/distribute
    * ☐ [pip_requirements.txt]()
        * ipython, simplejson, boto, fabric
* ☐ rvm/bundle
    * ☐ [Gemfile]()
        * nanoc3
* ☐ service | 服務
    * ☐ [nginx](${HOME}/nginx/) [pub]
    * ☐ [cgi](${HOME}/etc/spawn-fcgi) [pub]
    * ☐ [mysql](${HOME}/my.cnf) [pub]
    * ☐ [mysql client](${HOME}/.my.cnf) [pri]

## app | 應用

* ☐ iterm2
* ☐ tmux
    * ☐ gem install teamocil
    * ☐ [${HOME}/.teamocil]()
* ☐ brew
    * ☐ [install_brew.sh]()
    * ☐ [brew_pack.txt]()
          ec2-api-tools ec2-ami-tools dict netcat nethack
          sysstat dstat ntop iotop screen nginx 
          git curl rsync openssl python26 w3m elinks wget
          tree mutt dict spawn-fcgi autossh
          mysql
        
* ☐ yum
    * ☐ install_yum_repo
    * ☐ [pip_requirements.txt]()
          sysstat dstat ntop iotop screen nginx 
          git-all rsync openssl python26 w3m elinks wget
          tree mutt dict spawn-fcgi autossh
          mysql
    * ☐ yum cleanup

## operation | 运营
* ☐ backup | 备份
    * ☐ public data | 公共数据
        * ☐ public repos: github
        * ☐ databases
        * ☐ dot files
        
    * ☐ private data | 个人数据
        * ☐ ssh keys
        * ☐ mac os keychains
        * ☐ personal repos: bitbucket
        * ☐ databases
        * ☐ dot files

* ☐ 


## The Ref | 參考文檔
