---
created_at: 2013-11-06
excerpt: a basic introduction of gollum/nginx
kind: article
publish: true
tags: [dev, tools, doc]
title: "quickly starting gollum wiki behind nginx"
---
gollum wiki 介绍

    gem install gollum
    (rbenv rehash)
    mkdir -p /data/gollum_wiki
    cd /data/gollum_wiki
    git init

* gollum /data/gollum_wiki --port 4567

访问 localhost:5467 就可以开始编辑了！

可以加个auth.rb 密码登录

    # auth.rb
    module Precious
      class App < Sinatra::Base
        use Rack::Auth::Basic, "not public Area" do |username, password|
          [username, password] == ['adminname', '|------|']
        end
      end
    end

* gollum /data/gollum_wiki --port 4567 --config auth.rb

nginx端口转发

    server {
        listen   192.168.16.21:80;
        server_name 192.168.16.21 gollum.me.localhost;

        access_log  /var/log/nginx/gw.access.log;
        error_log   /var/log/nginx/gw.error.log info;

        location /{
            proxy_pass   http://127.0.0.1:4568/;
            proxy_redirect off;
            proxy_set_header   Host              $host;
            proxy_set_header   X-Real-IP         $remote_addr;
            proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
            proxy_set_header   X-Forwarded-Proto $scheme;
        }
    }


更多信息

* 语法帮助 https://github.com/gollum/
* git库 https://github.com/gollum/gollum/wiki

## The End | 文檔終點

