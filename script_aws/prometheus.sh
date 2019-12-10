#!/bin/bash

echo "installing prometheus..."
sudo groupadd --system prometheus
sudo useradd --no-create-home -s /sbin/nologin --system -g prometheus prometheus
sudo mkdir -p /var/lib/prometheus
sudo mkdir -p /etc/prometheus
sudo mkdir -p /etc/prometheus/rules
sudo mkdir -p /etc/prometheus/rules.d
sudo mkdir -p /etc/prometheus/files_sd
#for i in 'rules' 'rules.d' 'files_sd'; do sudo mkdir -p '/etc/prometheus/${i}'; done
mkdir -p /tmp/prometheus && cd /tmp/prometheus
curl -s https://api.github.com/repos/prometheus/prometheus/releases/latest \
  | grep browser_download_url \
  | grep linux-amd64 \
  | cut -d '"' -f 4 \
  | wget -qi -
tar xvf prometheus*.tar.gz 1> /dev/null 2> /dev/null
cd prometheus*/
sudo mv prometheus promtool /usr/local/bin/
sudo mv -n consoles/ console_libraries/ /etc/prometheus/
cd ~/
#sudo mv prometheus.yml  /etc/prometheus/prometheus.yml
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
tar -xvf node_exporter*.tar.gz
rm node_exporter-0.18.1.linux-amd64.tar.gz
cd  node_exporter*/
sudo systemctl stop node_exporter
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
