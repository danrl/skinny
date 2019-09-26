#!/bin/bash -x

sudo systemctl restart skinny@catbus.service
sudo systemctl restart skinny@kanta.service
sudo systemctl restart skinny@mei.service
sudo systemctl restart skinny@satsuki.service
sudo systemctl restart skinny@totoro.service

sleep 5

for I in $(seq 1 9)
do
    skinnyctl acquire beaver
done
