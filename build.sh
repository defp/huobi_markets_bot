#!/bin/bash

git pull origin master
make
supervisorctl restart huobi_bot