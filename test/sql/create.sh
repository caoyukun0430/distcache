#!/bin/bash

# MySQL 连接信息
USER="root"
PASS="1234"
HOST="127.0.0.1"
PORT="3306"
DBNAME="distcache"

# use mysql_config_editor to store your MySQL username and password
# this cmd only needs to be executed once 
# mysql_config_editor set --login-path=local --host=$HOST --user=$USER --password

# DB_EXISTS value:
# 0: Indicates that grep found a match (the database exists).
# 1: Indicates that grep did not find a match (the database does not exist).
DB_EXISTS=$(mysql --login-path=local -h$HOST -P$PORT -e "SHOW DATABASES LIKE '$DBNAME';" | grep "$DBNAME" > /dev/null; echo "$?")

# 如果数据库不存在，则创建它
if [ $DB_EXISTS -ne 0 ]; then
  mysql --login-path=local -u root -h$HOST -P$PORT -e "CREATE DATABASE $DBNAME;"
  echo "Database $DBNAME created successfully"
else
  echo "Database $DBNAME exists already, no need to create"
fi