#!/bin/sh

rsync \
  --partial --progress --times \
  conrad@conrad-thinkstation-s30:go/src/fknsrs.biz/p/ytmusic/database.db database.db
