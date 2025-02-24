---
created_at: 2012-07-06
excerpt: my own documentation for standard markdown syntax
kind: article
publish: true
tags: [doc, md, perm]
title: "markdown syntax"
---

### Literal | 一些背景
* created by: John Gruber & Aaron Swartz
* target: easy to read AND easy to write plain text format

### Syntax | 語法

* Headings - hashes before text
        ### with hash made this my heading

* Paragraphs - no indent
        without any indent this is just a paragraph

* Paragraphs - two spaces to end the paragraphs
_colon_ after paragraphs doesnt work with my version
        here we have two spaces hidden behind
        so this is two paragraphs

* Lists - bullted with \* and enumerated with number
        * my list item 1
        * my list item 2
            * sublist item indent with four spaces

* Code span - with backticks
        `<code me blue>`
`<code me blue>`

* Code block - formatted with four spaces prefixed
        In [2]: fabric?
        Type:       module

        Base Class: <type 'module'>
        String Form:<module 'fabric' from '/Users/arbeitandy/.virtualenvs/dailysa/lib/python2.7/site-packages/fabric/__init__.pyc'>

* Blockquotes
        > this would flow everthing like this: Where past and future are gathered. Neither movement from nor towards,

* Horizontal lines - triple stars, dashes
        ---
        * * *
        - - - -

* Manual line breaks - two or more spaces hidden behind.
        i made a line break here /spaces/
        with two spaces

* Phrase Emphasis - star or underscore
        *star me to be italic*
        ** double star me to be bold **

* Linkiefy - squarebracket the notation and bracket the link
        short: [home](http://pezware.com)
        full: [home](http://pezware.com "Title: HOME")
        use reference style: this is [home][linkid] yes it is
        ref(put it lateR): [linkid]: http://pezware.com "Title: HOME"

[linkid]: http://pezware.com "Title: HOME"

* Inline image - almost same as linkie with an "!" in front
        short: ![i am image](/files/img/tuxs.jpg)
        full: ![alt shelter](/files/img/tuxs.jpg "Gimme shelter")
        ref style: [thisimgid]: /files/img/tuxs.jpg "Title"
        use the ref: this is ![an img][thisimid] yes it is

this is this ![alt][thisimg] king!


* Shorthands to use linkies
        autolink: <http://example.com/>

        [google]: http://google.com/        "Google"
        [yahoo]:  http://search.yahoo.com/  "Yahoo Search"
        [msn]:    http://search.msn.com/    "MSN Search"







## The Ref | 參考文檔

* [original markdown syntax [EN]][ref1]
* [wikipedia markdown syntax [EN]][ref2]


[ref1]: http://daringfireball.net/projects/markdown/syntax/ "original markdown sytax"
[ref2]: http://en.wikipedia.org/wiki/Markdown#Syntax_examples "syntax on wiki



## current status: DONE
### The End | 文檔終點

back to [home](http://pezware.com)
[thisimg]: /files/img/tuxs.jpg "Title"
