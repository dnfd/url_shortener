# URL Shortener

### How to Run

#### Launch MySQL

```sh
docker-compose -f configs/compose.yml up -Vd
```

#### Run Server

```sh
go run ./cmd/server
```

#### Register new URL

```sh
curl -X POST -i localhost:8080/urls/new --data '{"url": "https://www.google.com"}'
```

#### Try Redirect

```sh
curl -i localhost:8080/returned_url
```

### Coverage

```sh
ok  	github.com/dnfd/url_shortener/cmd/server	0.011s	coverage: 68.8% of statements
ok  	github.com/dnfd/url_shortener/internal/urlconverter	(cached)	coverage: 76.9% of statements
```

### Benchmark

```sh
hey -n 100000 -disable-redirects http://localhost:8080/0

Summary:
  Total:	2.1570 secs
  Slowest:	0.0533 secs
  Fastest:	0.0001 secs
  Average:	0.0011 secs
  Requests/sec:	46360.0945


Response time histogram:
  0.000 [1]	|
  0.005 [98467]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.011 [1225]	|
  0.016 [217]	|
  0.021 [37]	|
  0.027 [16]	|
  0.032 [17]	|
  0.037 [9]	|
  0.043 [3]	|
  0.048 [6]	|
  0.053 [2]	|


Latency distribution:
  10% in 0.0003 secs
  25% in 0.0005 secs
  50% in 0.0008 secs
  75% in 0.0012 secs
  90% in 0.0019 secs
  95% in 0.0028 secs
  99% in 0.0067 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0001 secs, 0.0533 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0350 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0237 secs
  resp wait:	0.0007 secs, 0.0000 secs, 0.0207 secs
  resp read:	0.0002 secs, 0.0000 secs, 0.0239 secs

Status code distribution:
  [302]	100000 responses
```
