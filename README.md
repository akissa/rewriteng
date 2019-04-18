# rewriteng

## Name

*rewriteng* - New generation rewrite plugin.

## Description

The rewriteng plugin rewrites queries and responses

## Syntax

~~~
rewriteng CLASS RR-TYPE TYPE FROM-DOMAIN TO-DOMAIN {
    answer [name|data] exact|prefix|suffix|substring|regex STRING STRING
    additional [name|data] exact|prefix|suffix|substring|regex STRING STRING
    authority [name|data] exact|prefix|suffix|substring|regex STRING STRING
}
~~~

* **CLASS** the query class (usually IN or ANY).
* **RR-TYPE** the query type (A, PTR, ... can be ANY to match all types).
* **TYPE** the match type, i.e. exact, substring, etc., triggers re-write:
* **FROM-DOMAIN** the domain to rewrite
* **TO-DOMAIN** the domain to rewrite to

### Rules

The rule syntax is as follows:

~~~
# rule type                 rr part     match type                          from   to
answer|additional|authority [name|data] exact|prefix|suffix|substring|regex STRING STRING
~~~

#### Rule type

The following rule types are supported:

* **answer**: rewrites answers, atleast one answer is required multiple rules are allowed
* **additional**: rewrites the additional section, these are optional and multiple rules are allowed
* **authority**: rewrites the authority section, these are optional and multiple rules are allowed

#### RR part

The following RR parts are supported:

* **name**: rewrites the name part
* **data**: rewrites the data part
* **both**: rewrites both the name and data parts

If the RR part is omitted, the `name` RR part is assumed.

#### Match type

The match type, i.e. `exact`, `substring`, etc., triggers re-write:

* **exact** (default): on exact match of the name in the question section of a request
* **substring**: on a partial match of the name in the question section of a request
* **prefix**: when the name begins with the matching string
* **suffix**: when the name ends with the matching string
* **regex**: when the name in the question section of a request matches a regular expression

## Examples

The following rewrites queries to `x.example.com` to `x.yahoo.com`, it also rewrites the
authority and the additional sections.

~~~ corefile
.:5300 {
    log
    bind 127.0.0.1
    forward . 192.168.1.2
    rewriteng IN ANY suffix example.com yahoo.com {
        answer regex (.*)\.yahoo\.com {1}.example.com
        answer data regex (.*)\.yahoo\.com {1}.example.com
        authority suffix yahoo.com. example.com.
        authority data substring yahoo. example.
        additional suffix yahoo.com. example.com.
        additional data substring 68.142.254.15 192.168.1.2
        additional data substring 68.180.130.15 192.168.1.2
    }
}
~~~

The normal output without rewriting is as follows:

```
$ dig www.yahoo.com

; <<>> DiG 9.8.3-P1 <<>> www.yahoo.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 27720
;; flags: qr rd ra; QUERY: 1, ANSWER: 3, AUTHORITY: 4, ADDITIONAL: 2

;; QUESTION SECTION:
;www.yahoo.com.			IN	A

;; ANSWER SECTION:
www.yahoo.com.		1800	IN	CNAME	atsv2-fp-shed.wg1.b.yahoo.com.
atsv2-fp-shed.wg1.b.yahoo.com. 60 IN	A	87.248.98.7
atsv2-fp-shed.wg1.b.yahoo.com. 60 IN	A	87.248.98.8

;; AUTHORITY SECTION:
wg1.b.yahoo.com.	87717	IN	NS	yf3.a1.b.yahoo.net.
wg1.b.yahoo.com.	87717	IN	NS	yf2.yahoo.com.
wg1.b.yahoo.com.	87717	IN	NS	yf4.a1.b.yahoo.net.
wg1.b.yahoo.com.	87717	IN	NS	yf1.yahoo.com.

;; ADDITIONAL SECTION:
yf1.yahoo.com.		1317	IN	A	68.142.254.15
yf2.yahoo.com.		1317	IN	A	68.180.130.15

;; Query time: 22 msec
;; SERVER: 192.168.1.2#53(192.168.1.2)
;; WHEN: Thu Apr 18 09:43:21 2019
;; MSG SIZE  rcvd: 215

```

The rewrite of `www.example.com` to `www.yahoo.com` outputs the following:

```
$ dig www.example.com @127.0.0.1 -p 5300

; <<>> DiG 9.8.3-P1 <<>> www.example.com @127.0.0.1 -p 5300
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 12839
;; flags: qr rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 4, ADDITIONAL: 2

;; QUESTION SECTION:
;www.example.com.		IN	A

;; ANSWER SECTION:
www.example.com.	1662	IN	CNAME	atsv2-fp-shed.wg1.b.example.com.
atsv2-fp-shed.wg1.b.example.com. 60 IN	A	87.248.98.7

;; AUTHORITY SECTION:
wg1.b.example.com.	87579	IN	NS	yf2.example.com.
wg1.b.example.com.	87579	IN	NS	yf1.example.com.
wg1.b.example.com.	87579	IN	NS	yf3.a1.b.example.net.
wg1.b.example.com.	87579	IN	NS	yf4.a1.b.example.net.

;; ADDITIONAL SECTION:
yf1.example.com.	1179	IN	A	192.168.1.2
yf2.example.com.	1179	IN	A	192.168.1.2

;; Query time: 10 msec
;; SERVER: 127.0.0.1#5300(127.0.0.1)
;; WHEN: Thu Apr 18 09:45:39 2019
;; MSG SIZE  rcvd: 396

```

## Also See

See the original rewrite plugin it is used as the basis for this plugin.
