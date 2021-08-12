#!/bin/bash

DEST_DIR='./build/dist'

aliyuncli --region-id cn-beijing oss-upload --acl public-read \
    yunioniso 'rpms/telegraf/' $DEST_DIR
