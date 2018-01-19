#!/bin/bash

git pull origin master
go build .
supervisorctl restart huobi_bot