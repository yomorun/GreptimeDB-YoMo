# GreptimeDB x Yomo

This repository demonstrates how GreptimeDB can be used to transport data using YoMo.

1. install GreptimeDB (The example use docker, more install way see: [greptime docs](https://docs.greptime.com/getting-started/installation/overview))

```bash
docker run -p 127.0.0.1:4000-4003:4000-4003 \
-v "$(pwd)/greptimedb:/tmp/greptimedb" \
--name greptime --rm \
greptime/greptimedb:v0.8.2 standalone start \
--http-addr 0.0.0.0:4000 \
--rpc-addr 0.0.0.0:4001 \
--mysql-addr 0.0.0.0:4002 \
--postgres-addr 0.0.0.0:4003
```

2. install YoMo

```bash
curl -fsSL https://get.yomo.run | sh
```

3. run yomo zipper

```bash
yomo serve -c config.yaml
```

4. run yomo sfn, sfn bridges GreptimeDB between YoMo

```bash
cd sfn && yomo run app.go
```

5. run yomo source, The source watches a file and writes the new content from the file to the zipper.

```bash
go run source/main.go -f metric.log
```

Now The source can write Line Protocol data to the `metric.log` file. The data written to the file will be transmitted over the YoMo network and written to the GreptimeDB.

1. We provide a bash script to write to `metric.log`.

```bash
bash metric_ingest.sh
```

7. check data be written to the GreptimeDB.

```bash
curl -X POST \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'sql=select * from monitor' \
http://localhost:4000/v1/sql?db=public
```
