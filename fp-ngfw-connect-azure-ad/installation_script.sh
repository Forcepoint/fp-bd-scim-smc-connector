#!/bin/bash

if [[ $# -ne 1 ]];then
  echo "pass your Forcepoint SMC internal ip address as a parameter."
  echo "For example: ./installation_script.sh 192.168.122.13"
  exit 1
fi

if [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
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
  sudo yum install make gcc -y
  wget https://www.openssl.org/source/openssl-1.1.1f.tar.gz
  tar -zxf openssl-1.1.1f.tar.gz && cd openssl-1.1.1f
  ./config
  make
  sudo make install
  sudo ln -s /usr/local/bin/openssl /usr/bin/openssl
  echo "/usr/local/lib64" > /etc/ld.so.conf.d/openssl.conf
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

  echo "install Nginx"
  sudo yum install epel-release -y
  yes | sudo yum install nginx -y
  sudo firewall-cmd --permanent --zone=public --add-service=http
  sudo firewall-cmd --permanent --zone=public --add-service=https
  sudo firewall-cmd --reload
  setsebool -P httpd_can_network_connect 1
  cp -r ./nginx/certs /etc/nginx
  cp -f ./nginx/nginx.conf /etc/nginx/nginx.conf
  cp -f ./nginx/conf.d/ssl.conf /etc/nginx/conf.d/
  mkdir /var/azure_smc
  chmod +x scim-smc smc-connector deployment
  cp ./scim-smc /var/azure_smc
  cp ./smc-connector /var/azure_smc
  cp ./scim.yml /var/azure_smc
  cp ./connector.yml /var/azure_smc
  cp ./forcepoint_scim.service /etc/systemd/system/
  cp ./smc_connector.service /etc/systemd/system/
  yum provides /usr/sbin/semanage
  yum install policycoreutils-python -y
  echo "$1 smc.com" >> /etc/hosts
  sudo systemctl enable nginx.service
  sudo systemctl enable forcepoint_scim.service
  sudo systemctl enable smc_connector.service
else
  echo "the passed value '$1' is not a valid ip address"
  exit 1
fi

