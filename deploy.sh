#!/bin/bash

BACKUP=0
if [ $# -eq 1 ]; then
	if [ $1 = "-b" ]; then
		BACKUP=1
	fi
fi

cd ~/go/src/github.com/tuken/nix/

if [ $BACKUP -eq 1 ]; then
	# nix.yyyymmdd形式のファイルが3個以上あれば一番古いファイルを削除
	count=$(ls -1 nix.[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9] 2>/dev/null | wc -l)
	if [ "$count" -ge 3 ]; then
		oldest=$(ls -1t nix.[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9] | tail -1)
		rm -f "$oldest"
	fi
	cp -p nix nix.`date '+%Y%m%d'`
fi

go tool gqlgen generate
go build -ldflags="-s -w" -trimpath
sudo systemctl restart nix.service
