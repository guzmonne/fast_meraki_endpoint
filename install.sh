#!/bin/bash

FAST_MERAKI_ENDPOINT_VERSION=${FAST_MERAKI_ENDPOINT_VERSION:-"0.0.3"}

echo "Creating user fast_meraki_endpoint"
sudo useradd fast_meraki_endpoint -s /sbin/nologin -M
echo "Moving service configuration to /lib/systemd/system/"
sudo mv ./fast_meraki_endpoint.service /lib/systemd/system/.
sudo chmod 755 /lib/systemd/system/fast_meraki_endpoint.service
echo "Downloading fast_meraki_endpoint binaries"
wget --quiet https://github.com/guzmonne/fast_meraki_endpoint/releases/download/$FAST_MERAKI_ENDPOINT_VERSION/fast_meraki_endpoint
sudo mv ./fast_meraki_endpoint /usr/bin/fast_meraki_endpoint
echo "Creating application folders"
sudo mkdir -p /srv/fast_meraki_endpoint
echo "Enable and start service"
sudo systemctl enable fast_meraki_endpoint.service
sudo systemctl start fast_meraki_endpoint.service