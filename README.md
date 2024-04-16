# celestia-node-exporter
otel collector with prometheus exporter

## Usage

### Add metric params to init command
```
celestia bridge start <Other Params> --metrics --metrics.tls=false
```

### celestia mainnet
```
cp .env.celestia .env
docker compose up -d
```

### mocha testnet
```
cp .env.mocha .env
docker compose up -d
```