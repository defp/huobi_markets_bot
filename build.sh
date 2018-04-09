#!/bin/bash

git pull origin master
gb build
supervisorctl restart huobi_bot
