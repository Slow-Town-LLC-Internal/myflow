---
created_at: 2013-02-19
excerpt: a basic introduction of ansible in Chinese
kind: article
publish: true
tags: [dev, tools]
title: "开始使用配置和发布管理 with ansible"
---

趁着假期研究了一下，发现ansible比以前讨论过的chef/puppet/capistrano
和现在用的fabric等同类方便太多了

* Advantages
    * 好处1. 直接使用ssh协议(可以选择用openssh或者python封装的paramiko)
    * client不需要安装软件或者开启端口。client需要安装一个python-simplejson包(提供json格式返回结果)就可以直接使用
    * 好处2. 适合从ad-hoc的紧急工作到有序部署的大型架构，
因为ansible将系统的配置作为数据使用yaml格式保存，
而不是python code，所以系统配置/架构可以非常清晰。
对于复杂度高的多server和复杂度低的幂等架构都可以快速部署
    * 好处3. ansible源于开发了cobble的Michael DeHaan, 他也是puppet团队成员，
对于配置管理/部署等工作非常熟悉，很靠谱。
    * 需要考虑的: 对于ssh keybased验证和sudo权限需要预先了解和部署
    * 需要纪录的: 即使使用 openssh连接，ansible也会忽略 ~/.ssh/config中的Port配置，必须在inventory文件定义中显示标注

* 特征
    * 特征1. 使用类似于chef的push发放配置
    * 特征2. ansible的配置管理叫做playbook
    * 特征3. 可以使用普通用户，加入 -s 参数或者相应配置在playbook文件中启用sudo授权

* 试用（假设 ~/.ssh/config 中已经有remote_webserver_01 配置ssh keybased连接)
    *  安装：
        * 使用python的virtualenv建立一个ansible环境 (可忽略although建议不)
                mkvirtualenv ansi_env
                workon ansi_env
        * 安装依赖的包
                pip install yolk paramiko jinja2 PyYAML
                yolk -l
        * 直接使用pip包，目前是1.0版本
                pip install ansible
        * 或者用git HEAD
                cd ~/git_working
                git clone https://github.com/ansible/ansible.git github.gen.ansible
                cd !$
                git checkout devel
                # source ./hacking/env-setup  如果要使用缺省系统路径，就执行这一步
    *  一个简单的ansible inventory文件
      *  cat ./etc_ansible_hosts_sample
                # -- ansible_hosts_sample 01 - webservers
                ansible_sudo=root
                ansible_ssh_port=225

                [rack01_phy]
                192.168.66.91
                192.168.66.209

                [rack01_vm]
                192.168.03.[201:206]

                [idc01_new_initing]
                192.168.06.44:23455

                [idc02_db]
                10.12.0.14:23455
                10.12.0.16:23456

                [aws_farm]
                aws_06:22 ansible_python_interpreter=/usr/bin/python26
      * 最基本的ansible操作
              # -i 指定使用的inventory文件
              ansible -i ./etc_ansible_hosts_sample --list-hosts all
              ansible -i ./etc_ansible_hosts_sample --list-hosts rack01_vm

              # ansible ping module会检查remote server是否有python的simplejson模块，
              # 如果没有，会报错。-c ssh 指定连接方式为openssh
              ansible -i ./etc_ansible_hosts_sample -c ssh -u remote_devuser idc02_db -m ping
              # 使用raw模式安装ansible依赖的python-simplejson包, -s 是sudo
              export JSON_PACKAGE='python-simplejson.x86_64';
              ansible -i ./etc_ansible_hosts_sample \
              -c ssh -u remote_devuser idc01_new_initing \
              -s -a "yum install -y $JSON_PACKAGE"

* ☀ ansible中文名字叫做安塞波
    * ☀ 文档: http://ansible.cc/docs/
    * ☀ FAQ: http://ansible.cc/faq.html
    * ☀ playbooks sample: https://github.com/mpdehaan/ansible-examples
    * reference links
      * [compare with puppet(英文)](http://www.mail-archive.com/puppet-users@googlegroups.com/msg33389.html)
      * [remote env in bashrc](https://github.com/ansible/ansible/issues/1221) # 远程环境变量问题应该不算是ansible的问题，不过最好提前考虑好

## The End | 文檔終點

