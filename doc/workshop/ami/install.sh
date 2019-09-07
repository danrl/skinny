#!/bin/bash -x
set -e

# install tools
sudo apt-get install -y git mc tree vim

# install go language
sudo apt-get install -y build-essential
(
    cd /tmp
    wget -q https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz
    tar -xf go1.12.6.linux-amd64.tar.gz
    sudo mv go /usr/local
)

# configure go language
echo "export GOROOT=/usr/local/go" >> ~/.profile
echo "export GOPATH=\$HOME/go" >> ~/.profile
echo "export PATH=\$GOPATH/bin:\$GOROOT/bin:/home/ubuntu/skinny/bin:\$PATH" >> ~/.profile
source ~/.profile

# validate go language installation
go version
go env

# install build tools: protobuf
(
    sudo apt-get install -y protobuf-compiler
    go get -d -u github.com/golang/protobuf/protoc-gen-go
    git -C "$GOPATH/src/github.com/golang/protobuf" checkout v1.2.0
    go install github.com/golang/protobuf/protoc-gen-go
)

# install build tools: mage
(
    go get -v -u -d github.com/magefile/mage
    cd "$GOPATH/src/github.com/magefile/mage"
    go run bootstrap.go
)

# download skinny source code
git clone --single-branch --branch workshop-configs https://github.com/danrl/skinny.git ~/skinny

# build skinny
(
    cd skinny
    go mod vendor
    mage test
    mage
)

# install systemd service template
sudo bash -c 'cat >/etc/systemd/system/skinny@.service << EOF
[Unit]
Description=Skinny Instance %I

[Service]
Type=simple
ExecStart=/home/ubuntu/skinny/bin/skinnyd --config=/home/ubuntu/skinny/doc/workshop/configs/%i.yml
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF'

# enable systemd services
sudo systemctl enable skinny@catbus.service
sudo systemctl enable skinny@kanta.service
sudo systemctl enable skinny@mei.service
sudo systemctl enable skinny@satsuki.service
sudo systemctl enable skinny@totoro.service

# start systemd services
sudo systemctl start skinny@catbus.service
sudo systemctl start skinny@kanta.service
sudo systemctl start skinny@mei.service
sudo systemctl start skinny@satsuki.service
sudo systemctl start skinny@totoro.service

# test quorum is working
sleep 30
/home/ubuntu/skinny/bin/skinnyctl --config=/home/ubuntu/skinny/doc/workshop/quorum.yml status
sleep 10
/home/ubuntu/skinny/bin/skinnyctl --config=/home/ubuntu/skinny/doc/workshop/quorum.yml acquire beaver
sleep 10
/home/ubuntu/skinny/bin/skinnyctl --config=/home/ubuntu/skinny/doc/workshop/quorum.yml status
