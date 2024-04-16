# celestia-bridge-exporter
celestia bridge node exporter

## Build
```
go build .
```

## Systemd config
```
[Unit]
Description=Celestia Bridge Exporter  
After=network.target  
  
[Service]  
User=<your-user> 
Group=<your-user>
Type=simple  
ExecStart=/home/<your-user>/celbridge-exporter --listen.port 8380 --endpoint http://localhost:26658 --p2p.network mocha
  
[Install]  
WantedBy=multi-user.target  
```

## Forked from
[CelestiaTools](https://github.com/Chainode/CelestiaTools)