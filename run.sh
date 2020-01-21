#! /bin/sh
date_time=`date +%Y%m%d%H%M%S`
echo $date_time
./Scraper

git config core.sshCommand 'ssh -i ~/.ssh/id_rsa_yangwenmai'

git add .
git commit -m "docs: update `date`"
git push origin master

