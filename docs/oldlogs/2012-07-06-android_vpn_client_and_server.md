---
created_at: 2012-07-06
excerpt: draft documentation over how to setup vpn server on aws ec2 vm
kind: article
publish: true
tags: [misc, doc]
title: "open vpn server on ec2 for android client"
---

## The Server | 服務端

* Env | 環境

        % amazon ec2 micro instance
        % CentOS 5.6
                % sudo yum install -y openswan ppp xl2tpd

* Conf | 配置

        % /etc/xl2tpd/xl2tpd.conf
        % /etc/ppp/options.xl2tpd

* Run | 運行



### The Client | 客戶端

* Conf | 配置

        %
        %

* Run | 運行



### The Ref | 參考文檔

* [android配置[中文]](http://www.luojs.com/2011/03/31/android-vpn-one/)
* [android和server配置[中文]](http://www.jfwhome.com/2012/01/29/ipsecl2tp-vpn-on-linode-ubuntu-server-for-iphoneandroid/)
* [server config [english]](http://wiki.nikoforge.org/L2TP/IPSec_VPN_Setup_on_Centos_6_\(64-bit\)_for_use_with_Android_ICS_and_iOS_5_Clients)
* [racoon server config[EN]](http://www.howtoforge.com/racoon_roadwarrior_vpn)
* [android配置pptpd[EN]](http://noiseandheat.com/blog/2012/01/vpn-with-amazon-ec2-or-saving-your-phone-from-promiscuity/)
* [strongswan in aws vps[EN]](http://wiki.strongswan.org/projects/strongswan/wiki/AwsVpc)



## The End | 文檔終點

