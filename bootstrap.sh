sudo install -m 644 prometheus/prometheus.yml /etc/prometheus/prometheus.yml
sudo install -d /etc/prometheus/targets
sudo install -m 644 prometheus/targets/*.yml /etc/prometheus/targets/

sudo install -d /etc/grafana/provisioning/datasources
sudo install -m 644 -o root -g grafana grafana/provisioning/datasources/prometheus.yml /etc/grafana/provisioning/datasources/
