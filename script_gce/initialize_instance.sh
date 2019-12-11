#!/bin/bash
#initialize google compute vm

echo "updating OS..."
sudo apt-get -q update 1> /dev/null #2> /dev/null
sudo apt-get -q -y upgrade 1> /dev/null

echo "installing git..."
sudo apt-get install git -y -q 1> /dev/null

echo "installing go..."
sudo apt-get install golang -y -q 1> /dev/null

echo "installing jq..."
sudo apt-get install jq -y -q 1> /dev/null

echo "installing project dependency..."
go get -u cloud.google.com/go/monitoring/apiv3
go get -u github.com/aws/aws-sdk-go
go get -u github.com/samuel/go-zookeeper/zk
go get github.com/influxdata/influxdb1-client/v2

echo "installing zookeeper"
sudo apt-get install default-jdk -y -q 1> /dev/null
sudo wget -q -nc https://www-us.apache.org/dist/zookeeper/zookeeper-3.5.6/apache-zookeeper-3.5.6-bin.tar.gz
sudo tar -xzf  apache-zookeeper-3.5.6-bin.tar.gz
sudo mv -n apache-zookeeper-3.5.6-bin ./zookeeper
sudo mkdir -p /var/lib/zookeeper


echo "installing project..."
cd ./go/src
sudo rm -rf progettoSDCC
git clone git@github.com:cesto93/progettoSDCC -q
mkdir -p ./progettoSDCC/configuration/generated
mkdir -p ./progettoSDCC/log
mkdir -p ./progettoSDCC/bin

echo "installing stackdriver-agent..."
curl -sSO https://dl.google.com/cloudagents/install-monitoring-agent.sh
sudo bash install-monitoring-agent.sh 1> /dev/null 2> /dev/null
rm install-monitoring-agent.sh

echo "installing prometheus..."
sudo groupadd --system prometheus
sudo useradd --no-create-home -s /sbin/nologin --system -g prometheus prometheus
sudo mkdir -p /var/lib/prometheus
for i in rules rules.d files_sd; do sudo mkdir -p /etc/prometheus/${i}; done
mkdir -p /tmp/prometheus && cd /tmp/prometheus
curl -s https://api.github.com/repos/prometheus/prometheus/releases/latest \
  | grep browser_download_url \
  | grep linux-amd64 \
  | cut -d '"' -f 4 \
  | wget -qi -
tar xvf prometheus*.tar.gz 1> /dev/null 2> /dev/null
cd prometheus*/
sudo mv prometheus promtool /usr/local/bin/
#sudo mv prometheus.yml  /etc/prometheus/prometheus.yml
sudo mv consoles/ console_libraries/ /etc/prometheus/
cd ~/
sudo mv prometheus.yml  /etc/prometheus/prometheus.yml
rm -rf /tmp/prometheus
sudo tee /etc/systemd/system/prometheus.service<<EOF 1> /dev/null 2> /dev/null

[Unit]
Description=Prometheus
Documentation=https://prometheus.io/docs/introduction/overview/
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/local/bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/var/lib/prometheus \
  --web.console.templates=/etc/prometheus/consoles \
  --web.console.libraries=/etc/prometheus/console_libraries \
  --web.listen-address=0.0.0.0:9090 \
  --web.external-url=

SyslogIdentifier=prometheus
Restart=always

[Install]
WantedBy=multi-user.target
EOF

for i in rules rules.d files_sd; do sudo chown -R prometheus:prometheus /etc/prometheus/${i}; done
for i in rules rules.d files_sd; do sudo chmod -R 775 /etc/prometheus/${i}; done
sudo chown -R prometheus:prometheus /var/lib/prometheus/
sudo systemctl daemon-reload
sudo systemctl start prometheus
sudo systemctl enable prometheus 1> /dev/null 2> /dev/null
echo "installing node_exporter..."
curl -s https://api.github.com/repos/prometheus/node_exporter/releases/latest \
| grep browser_download_url \
| grep linux-amd64 \
| cut -d '"' -f 4 \
| wget -qi -
tar -xvf node_exporter*.tar.gz 1> /dev/null 2> /dev/null
rm node_exporter-0.18.1.linux-amd64.tar.gz
cd  node_exporter*/
sudo cp node_exporter /usr/local/bin
sudo tee /etc/systemd/system/node_exporter.service <<EOF 1> /dev/null 2> /dev/null
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=default.target
EOF

sudo systemctl daemon-reload
sudo systemctl start node_exporter
sudo systemctl enable node_exporter 1> /dev/null 2> /dev/null
