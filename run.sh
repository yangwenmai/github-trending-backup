#! /bin/sh
date_time=`date "+%G-%m-%d %H:%M:%S"`
echo $date_time
./Scraper

git add .
git commit -m "docs: update Github trending in $date_time"
git push origin master

