---
created_at: 2012-07-27
excerpt: draft documentation for aws ec2 and how it works together with dev process
kind: article
publish: true
tags: [misc, doc]
title: "aws clouding"
---

## cloud usage outline

- 概念
        IAAS - Infrastructure as a Service
        "alice覺得房間一會兒變大，一會兒變小"
            - "xxx" by xxx
        "我發現負載一會兒變高，一會兒變低"
            -  SA team


    note 重點cover IAAS以架構為服務
    這也是amazon aws提供的所有服務中最精華的部分
    值回票價，
    和以前的 PAAS/SAAS (Platform/Software) 平台和
    軟件級別的雲相比，
    基礎架構級別的雲是基於對節點的操作架構的雲。
    "對節點操作架構" 可以理解為 "對vm虛擬主機的操作和架構"
    "可伸縮性"從代碼級別實現了。
    在我們開始談論"可伸縮性"之前，先簡單看一下提供
    和承諾提供這個特性的服務商－以及我們的研發流程需要為此
    做什麼。

- 服務商
        - 國外
        aws, heroku, openshift(by redhat), hpcloud, gogrid,softlayer
        - 國內
        盛大雲/ali雲/huawei雲
        新浪 IAAS>PAAS
    - 企業私有雲是否保留IAAS?
    - vmware miniSAE

    note 目前提供對節點操作的服務商其實只有aws最為全面。
    但是如果應用較小並且研發流程簡單，可以考慮從heroku
    開始－因為heroku是遵從IAAS的基石 - 12 factor的部署
    原則，在應用擴展時候可以輕易轉移到aws，並且其它較
    為嶄新和較小的服務商最好讓其它人去測試。

    note 遺憾的是，在國內目前還沒有全部滿足IAAS的雲服務
    沒有實際應用，看了feature介紹和文檔之後瞭解
    到的, 新浪SAE的服務誠意十足，雖然從IAAS轉向了
    PAAS模式提供公開服務，還是提供vmware miniSAE測試機
    說不定會在未來從PAAS再擴展回到IAAS服務

- IAAS的革新
        - need 12 factor
        - http://www.12factor.net/

    note 在研發層次需要逐步引進12 factor
    12 factor 是什麼？一個研發標準化的manifesto
    確保應用能做到獨立於platform運行，可以直接部署在類似於heroku這樣的PAAS/IAAS平台



- 雲在內網 | internal cloud

    note 現實不是那麼光明，雖然在內部使用了五年的虛擬
    服務，但實際上和國內的雲服務較為類似，仍然停留在
    主機頒發研發帳號階段。很多feature都還在等一下要
    討論到的"計劃"中。
    主要是手工的IAAS，內網IT服務的操作和維護過去
    不涉及研發，維護一套IAAS的接口添加成本較高。
    在未來會加入開源工具比如google的ganeti或者
    vagrant/vee 補充測試部門所需要的IAAS需求。

- from xen to kvm to virtualbox
    - compare 實施方式
    - tips of using them
    - +vagrant now!

    note xen比較穩定並且操作配置相當簡單
    但是更換到Centos6以後，我們自然
    切換到了kvm進行新的vm管理，實際
    運營和xen的vm系統沒有不同。
    計劃部分我們深入討論一下對vagrant
    的應用計劃。

- IT system and testing
    - 內網可以虛擬化
    - 網絡配置變得重要了
    - 配置管理變得重要了
    - 命令行和代碼都集成了雲管理

    note 網絡的虛擬化在內網操作涉及不多，
    因為內網已經是VLAN架構，不必要
    再添加更多的網絡資源調配
    配置管理chef/puppet還未能普及
    目前的bottomline使用git central repo
    linux監控在很久以前就用了集群概念了
    但是資源不能切分，ganglia可以累加統計
    命令行操作透過fabric和onall執行
    可以快速獲取所有主機信息.
    測試集群涉及到代碼和數據庫的霧化-then雲化
    我們還沒法做到類似12factor的平行stateless
    應用部署需求

- 雲在外網 | external cloud
    - 云存储、云安全、云网络优化
        - akamai
        - 亞馬遜 aws (我們目前只用過了這個雲)
    - compairing with hardbox
        - 持久硬件 vs 快銷硬件(一次性硬件 - 但是good for environmnet)

    note ec2 腳本就是免費的硬件管理框架
    google說硬件會fail，aws也並非100%靠譜，已經大小事故好幾次了。

            STOP 硬件全都變了
                    - break all dependencies on hardware (including network asset)
                    STOP 更換硬件變得如此容易 http://alestic.com/2011/02/ec2-move-hardware
                ec2-stop-instances $instanceid
                ec2-start-instances $instanceid
                [opt] ec2-associate-address --instance $instanceid $ipaddress
                    FIXME 修復軟件也不難
                ec2-detach-volume $volume
                ec2-attach-volume --instance $instance_a --device /dev/sda1 $volume
                ec2-start-instances $instance_a

                    INFO 還有免費的meta data管理
                wget -q -O - http://169.254.169.254/latest/meta-data/public-ipv4

    note aws的接口太強大了。它們提供各種功能的可能。
    不過根據我們之前強調的12 factor，我們的方案
    會集中在: 在系統報警時迅速啓動ami鏡像擴展前端
    在閑置過多的時候自動關閉部份前端。
    自動更換hang掉的硬件



- 從ec2到full aws
        - vpc雲一樣的網絡

    note 和aws的其它部分服務，例如s3, rds相比，整套aws提供了對於應用
    架構上的所有資源節點，包括"網絡資源", "cpu資源", "主機資源"
    和"存儲資源"的直接操作。這些操作通過aws的api提供
    直接的命令行，包括aws-iam-tools, ec2-api-tools, ec2-ami-tools
    第三方的工具包括boto(python)


- integration into the current workflow
    - 架構補充

    note 在研發階段確認
    基礎環境，代碼，持久數據(metadata)
    短生命週期數據(session)，長生命週期數據(log)
    運行環境變量
    這些數據和變量通過嚴格分離之後就可以透過各種
    配置管理進行部署了。

- 配置
    - Virtualizaion + Configuration Management?
    - vagrant [ruby+virtualbox+puppet|chef|shell]
- 擴展
- 伸縮 [主要在前段，以後可能也包括db]
- 集成實施 [目前架構|數據|deployment|監控系統]

    note puppet部署基礎環境
    chef/fabric部署代碼和負責不斷發佈
    memcached, mq保存短生命週期數據
    mysqld集群和日誌中心保存並且對長生命週期數據
    進行同步等待followup的數據分析處理

    note 快速部署中完全不涉及數據的改變。
    但是mysql 需要一個更強壯的伸縮方案（當然主要是伸/擴展)
    可以考慮通過分片集群


    note 集成測試方案還在補充中－ [...]

- 雲在未來
  - dev box
    - 使用virtualbox打造開發環境並分發給dev team
    - 保持vm的版本一致來控制研發環境一致
  - 可伸縮性
    - 前段伸縮
    - db伸縮
    - 使用ec2 命令和boto庫構造快速構建的硬件系統

    note 剛才討論的很多方案還是藍圖，再概覽一下藍圖的全圖

- 雲在別處
  - 比如instagram,
  - 比我們早做好了這些方案的第三方
    - rightscale, cloud42, bitnami, scalr, ylastic
  - 優秀的雲管理界面讓你們不需要SA了
    - http://www.madeiracloud.com/support/kb/ide-interface/canvas/apps
    - 當然你們還是會有SA，但是他們現在和dev一起工作研究架構整個stack
    - 並改善流程和架構
  - aws case study http://aws.amazon.com/solutions/case-studies/

    note rightscale提供一個付費服務的伸縮前段管理
    他們還有非常全面的文檔和整套開放源代碼的腳本

- 對於vm的物理硬件的區別和經驗體會
  - 雲硬件不靠譜系數
    - (軟件不可靠系數 + 硬件不可靠系數) * 人工不靠譜系數
    - 硬件成為SA可操作資源

    note 和新浪的@easy所做的ppt一樣，也面臨著如何將雲和12 factor加入到
    研發流程的問題。舊的研發習慣很難改變，應用的耦合度，如果能解
    早就解了。規範的執行還有windows的障礙。

    note 但最終 希望能夠更有效地控制硬件資源，而不是按照最大容量x2的
    硬計劃購買大批量主機並且在晚上聽他們轟隆隆無所事事地空轉。
    所以無論從內網將資源變成雲，還是在對外服務中充分利用已經很優秀
    的雲方案，都是一種出路。

- 需要更多計算
    - 計算容量 - 並且快速
    - 計算[價格，CAPEX，預算]
    - 硬件和操作即時影響CAPEX

    note 雲計算設計到更多的計算和規劃。

    note 現在我們又多了一個成本核算層！
    並且這個成本由應用的performance直接控制
    雖然很快很靈活，但是是否有時就讓網站默默當掉算了呢？
    樂觀地考慮，集成成本核算是一種解決方案。

- 外面的雲都有自己的特點
    - amazon 全套工具和腳本
    - 從workflow集成開始S3 就和EC2 不一樣了

    note 這部分代碼可以很簡單，也可以外包（給比如rightscale這樣的aws api
    集成商，也可以自己做。

- vm給研發流程的正面影響
    - 強制性的快速deployment
    - 強制性的配置自動管理
    - 強制性的監控

    note 如之前強調的，可伸縮性只是運用雲的一個成本
    上的改進（說不定還無法改進)，IAAS的優勢在于
    剛才已經講的，提供和理想中的12 factor模式更
    契合的工作流。
    當然12 factor只是個manifesto，它並非研發標準，
    換句話說，如果代碼簡單，負載平衡，金錢資源
    充足，也完全不一定使用雲或者這些條例流程。

- 虛擬化的焦慮
    - 以上強制執行的系統都需要時間完善
    - 價格可能比預算高，如果cpu使用率超過計劃，可能會被賬單嚇一跳
    - 難以從業務角度確認哪些數據是必須持久化
    - 因為需要workflow改變，實施困難和緩慢。需要明確policy between dev/sa
    - 沒有顯著成效 (價格並未降低)
    - auto-provision 也會出錯。
    - 市場還未形成，沒有賣家也沒有買家

    note 有焦慮也有動力，總體來說，IAAS雲是一個很讓
    人激動並且有充分生命力的改變，希望能和
    服務商一起進步和成熟發展。也期待雲服
    務創建更大的市場和更多可能。

##    The Ref | 參考文檔


#    The End | 文檔終點

