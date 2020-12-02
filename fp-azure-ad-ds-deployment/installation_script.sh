#!/bin/bash

echo "install Python3.."
sudo yum install -y https://centos7.iuscommunity.org/ius-release.rpm
sudo yum install -y python36u python36u-libs python36u-devel python36u-pip

echo "installing Golang.."
sudo yum install wget -y
wget https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
sudo tar -zxvf go1.14.1.linux-amd64.tar.gz -C /usr/local
echo "export GOROOT=/usr/local/go" | sudo tee -a /etc/profile
echo "export PATH=$PATH:/usr/local/go/bin" | sudo tee -a /etc/profile
source /etc/profile

echo "Upgrade openssl .."
yum install make gcc perl pcre-devel zlib-devel -y
wget https://ftp.openssl.org/source/old/1.1.1/openssl-1.1.1.tar.gz
tar -zxf openssl-1.1.1.tar.gz && cd  openssl-1.1.1/
./config --prefix=/usr --openssldir=/etc/ssl --libdir=lib no-shared zlib-dynamic
make
sudo make install

export LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
echo "export LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64" >> ~/.bashrc
source ~/.bashrc
sudo ldconfig
cd ..
sleep 10

echo "installing Azure Cli..."
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
sudo sh -c 'echo -e "[azure-cli]
name=Azure CLI
baseurl=https://packages.microsoft.com/yumrepos/azure-cli
enabled=1
gpgcheck=1
gpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/azure-cli.repo'
sudo yum install azure-cli -y
chmod +x  deployment
