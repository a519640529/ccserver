#!/bin/bash
DUMP=/home/mongodb/bin/mongodump
#备份文件临时目录
OUT_DIR=/home/mongodb/backup/tmp
#备份文件正式目录
TAR_DIR=/home/mongodb/backup
#备份文件将以备份时间保存
DATE=`date +%Y_%m_%d_%H_%M_%S`
#数据库操作员
DB_USER=hjgame
#密码
DB_PASS=hj888
#保留最新14天的备份
DAYS=14
#备份文件命名格式
TAR_BAK="mongodb_bak_$DATE.tar.gz"
#创建文件夹
cd $OUT_DIR
#清空临时目录
rm -rf $OUT_DIR/*
#创建本次备份的文件夹
mkdir -p $OUT_DIR/$DATE
#执行备份命令
$DUMP --host 192.168.1.91 --port 3017 -u $DB_USER -p $DB_PASS -o $OUT_DIR/$DATE
#将备份文件打包放入正式目录
tar -zcvf $TAR_DIR/$TAR_BAK $OUT_DIR/$DATE
#删除14天前的旧备份
find $TAR_DIR/ -mtime +$DAYS -delete

