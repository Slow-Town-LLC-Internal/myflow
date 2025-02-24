---
created_at: 2012-10-24
excerpt: draft documentation over monitor system for system admins
kind: article
publish: true
tags: [misc, doc]
title: "monitor systems"
---

## The Server | 服務端

* Env | 環境
    * monitor by obfuscurity
        * Good = rrdtool, collectd, graphite
        * Adequate = reconnoiter, munin, opentsdb, ganglia
        * Bad = zenoss, zabbix, groundworks

- graphite [plot]
  * http://graphite.wikidot.com
  * [doc] http://www.aosabook.org/en/graphite.html
- collectd [sensor]
  * http://collectd.org
- coda hale`s metric lib @github [sensor]
  * https://github.com/codahale/metrics
- etsy statsd @github [aggregator]
  * https://github.com/etsy/statsd
- etsy logster @github [extract & aggregate]
  * https://github.com/etsy/logster
- logstash [log relay & aggregate]
  * http://logstash.net
- riemann monitor [external i/o sensor etc]
  * http://aphyr.github.com/riemann/
- sensu [external i/o distributed]
  * https://github.com/sensu
- heroku umpire @github [query]
  * https://github.com/heroku/umpire
- comsat @github [notifier]
  * https://github.com/asenchi/comsat
- kibana [dashboard for logstash]
  * http://kibana.org
  * arch doc: http://kibana.org/infrastructure.html
- tasseo [dashboard for logstash/realtime]
  * https://github.com/obfuscurity/tasseo
  * http://obfuscurity.com/2012/05/The-Story-Behind-Tasseo
- descartes [dashboard for logstash/convention]
  * https://github.com/obfuscurity/descartes


* via [original source - obfuscurity][ob_src]

### The Ref | 參考文檔

[ob_src]: https://speakerdeck.com/u/obfuscurity/p/the-state-of-open-source-monitoring "obfuscurity keynote"

## The End | 文檔終點

