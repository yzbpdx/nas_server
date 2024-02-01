#!/bin/bash

service mysql start
service redis-server start

mysql < /server/script.sql

/server/nas_server
