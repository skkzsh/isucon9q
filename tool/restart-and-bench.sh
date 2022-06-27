#!/bin/bash
set -euvx

# sudo truncate -s 0 -c /var/log/nginx/access.log
sudo truncate -s 0 -c /var/log/mysql/mysql-slow.log
# mysqladmin flush-logs

cd /home/isucon/isucari/webapp/go
make
sudo systemctl restart isucari.golang

sudo systemctl restart mysql
sudo systemctl restart nginx

cd ~/isucari
./bin/benchmarker

# sudo cat /var/log/nginx/access.log | \
#     alp json --sort avg -r -m '^/api/announcements/[0-9A-Z]+$','^/api/courses/[0-9A-Z]+$','^/api/courses/[0-9A-Z]+/status$','^/api/courses/[0-9A-Z]+/classes$','^/api/courses/[0-9A-Z]+/classes/[0-9A-Z]+/assignments$','^/api/courses/[0-9A-Z]+/classes/[0-9A-Z]+/assignments/export$','^/api/courses/[0-9A-Z]+/classes/[0-9A-Z]+/assignments/scores$'

# sudo mysqldumpslow /var/log/mysql/mysql-slow.log
# sudo pt-query-digest /var/log/mysql/mysql-slow.log

# go tool pprof -http=:10060 http://localhost:6060/debug/pprof/profile

