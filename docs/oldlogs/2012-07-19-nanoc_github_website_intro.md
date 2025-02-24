---
created_at: 2012-07-19
excerpt: draft about building a website using nanoc and other pub tools then maint it.
kind: article
publish: true
tags: [draft, doc, checklist]
title: "使用nanoc和github架設網站checklist"
symbol: ☐ ☑ ☒ 
---

## summary | 簡述

    文檔提供一個使用第三方服務建立簡單網站的checklist
    其中OPT標記的部分都可以省略。

## service | 服務

* ☐ domain via registrar | 域名註冊 [$][OPT]
    * [namecheap](http://www.namecheap.com)
    * [networksolutions](http://www.networksolutions.com)
    * 或者通過 [Google App](https://www.google.com/a/cpanel/standard/new?hl=en) 
    * 或者使用github.com page功能


* ☐ server | 服務器 [opt][$]
    * [linode](http://www.linode.com/)
    * [aws ec2](http://aws.amazon.com/ec2/)

* ☐ Google Service | Google 服務  
    * ☐ [Gmail](https://mail.google.com) [free]
    * ☐ [Google Analytics](http://www.google.com/analytics/) [free]
    * ☐ [Google Webmasters](http://www.google.com/webmasters/) [free]
    * ☐ [Google App](https://www.google.com/a/cpanel/standard/new?hl=en) [free google app]
    * ☐ [Google App Paid](http://www.google.com/enterprise/apps/business/pricing.html) [$]

* ☐ github | github 放置代碼/網站
    * ☐ [github](https://github.com) [free]

* ☐ DNS | 域名配置 [free]  
    * ☐ A紀錄 domain_name.com www.domain_name.com
    * ☐ MX紀錄 使用[Google App收郵件的配置](http://support.google.com/a/bin/answer.py?hl=en&answer=140034)
    * ☐ TXT紀錄 使用Google 服務的部分verification (case by case)
    * ☐ [sitelutions](http://sitelutions.com/) (比較靈活的第三方域名解析) [free][OPT]
        

* ☐ disqus | disqus 評論管理
    * ☐ [disqus](http://disqus.com/) [free]

## Init | 初始化操作

    這個步驟只操作一次
* ☐ fork [nanoc3_blog 代碼](https://github.com/mgutz/nanoc3_blog) 到本地的github帳號
* ☐ clone 到本地機器以備編輯和調試
* ☐ 配置nanoc3_blog 代碼加入網站信息
    * ☐ website name 網站名稱
    * ☐ disqus 帳號
    * ☐ google analytics 帳號
* ☐ 在服務器上的 [OPT]
    * ☐ 服務器基礎配置
    * ☐ nginx 配置
    * ☐ ruby/rvm 配置


## Maintain | 運營操作

    以後可以長期操作的步驟
* ☐ 在本地進行內容更新和調試 [OPT]
* ☐ 透過Google Analytics/Webmaster進行訪問數據查看
* ☐ 透過Github發佈內容更新
    
## Debug | 問題處理和其它

    需要根據本地配置時常檢查的問題/部分問題自行補充
* ☐ domain name過期續費 YY-MM-DD ______
* ☐ 設置gmail郵件備份
* ☐ 其它付費服務的續費
    * ☐ ______
    * ☐ ______
* ☐ 其它延期服務的延期和有限制的服務的上限檢查
    * ☐ ______
    * ☐ ______
* ☐ 如果包括dev流程和/或自行運營 [OPT]
    * ☐ 研發基本流程 [需求/開發/測試/發佈]
    * ☐ 運營基本點 [配置管理/健康和安全監控/容量評估/CAPEX預算]

## The Ref | 參考文檔
