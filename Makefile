.PHONY: install journal build clean

build: clean
	go build

clean:
	rm -f fast_meraki_endpoint

install:
	echo "Building the new version of fast_meraki_endpoint" ;\
	go build ;\
	echo "Stoping the fast_meraki_endpoint service" ;\
	sudo systemctl stop fast_meraki_endpoint.service ;\
	sudo cp ./fast_meraki_endpoint.service /lib/systemd/system/. ;\
	sudo chmod 755 /lib/systemd/system/fast_meraki_endpoint.service ;\
	sudo cp ./fast_meraki_endpoint /usr/bin/fast_meraki_endpoint ;\
	sudo mkdir -p /srv/fast_meraki_endpoint ;\
	echo "Restarting service" ;\
	sudo useradd fast_meraki_endpoint -s /sbin/nologin -M ;\
	sudo systemctl enable fast_meraki_endpoint.service ;\
	sudo systemctl start fast_meraki_endpoint.service

journal:
	sudo journalctl -f -u fast_meraki_endpoint

pprof:
	go tool pprof --web ~/pprof/mem.pprof