#!/bin/sh
/usr/sbin/sshd -D &
./loggen -file /var/log/hadoop/hadoop.log -node "$1" -msgpersec 5 -payload-gen constant