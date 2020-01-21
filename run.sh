#! /bin/sh
date_time=`date +%Y%m%d%H%M%S`
echo $date_time
./Scraper

git add .
git commit -m "docs: update `date`"
git push origin master

