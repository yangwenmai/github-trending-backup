#! /bin/sh
date_time=`date "+%G-%m-%d %H:%M:%S"`
echo $date_time
./Scraper

git add .
git commit -m "docs: update $date_time trending"
#git push origin master

