# ping_exporter

This exporter will do ping and tcping to serveral hosts and then collect the latency of the ping and tcping result.

config file as below:

```json
{
    "port": 8037,                       // listen port
    "max_latency": 300,                 // if the latency larger this, set this number as the ping result.
    "tcping": [                         // do tcping host list
        {
            "name": "CT",               // prometheus metrics label name
            "host": "www.189.cn",       // prometheus metrics label host
            "isIPv6": false,            // whether enable ipv6
            "port": 443                 // tcp port to do tcping
        },
        {
            "name": "CM",
            "host": "www.10086.cn",
            "isIPv6": false,
            "port": 443
        },
        {
            "name": "CU",
            "host": "www.chinaunicom.com.cn",
            "isIPv6": false,
            "port": 443
        }
    ],
    "ping": []                          // not complement yet
}
```

### how to run

clone this repository, build it.

```shell
go build main.go
./main -c /path/to/config.json
```

metrics be like

```shell
$ curl http://127.0.0.1:8037/metrics
tcping_latency{host="www.189.cn",isv6="false",name="CT"} 11
tcping_latency{host="www.chinaunicom.com.cn",isv6="false",name="CU"} 7
tcping_latency{host="www.10086.cn",isv6="false",name="CM"} 20
```