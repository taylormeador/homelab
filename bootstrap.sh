sudo install -m 644 prometheus/prometheus.yml /etc/prometheus/prometheus.yml
sudo install -m 644 prometheus/blackbox.yml /etc/prometheus/blackbox.yml
sudo install -d /etc/prometheus/targets
sudo install -m 644 prometheus/targets/*.yml /etc/prometheus/targets/

sudo install -d /etc/grafana/provisioning/datasources
sudo install -m 644 -o root -g grafana grafana/provisioning/datasources/prometheus.yml /etc/grafana/provisioning/datasources/

sudo install -d /etc/grafana/provisioning/dashboards
sudo install -m 644 -o root -g grafana grafana/provisioning/dashboards/homelab.yml /etc/grafana/provisioning/dashboards/

sudo install -d /var/lib/grafana/dashboards
sudo install -m 644 grafana/dashboards/*.json /var/lib/grafana/dashboards/

sudo systemctl restart prometheus prometheus-blackbox-exporter grafana-server
