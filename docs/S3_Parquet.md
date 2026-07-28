# S3 Parquet + flowBuffer (Phase 1 contract)

Operator-generated configuration aligned with landed flowlogs-pipeline and console plugin.

## FlowCollector API (summary)

```yaml
spec:
  loki:
    enable: false
  processor:
    flowBuffer:
      enable: true          # default: !loki.enable
      maxEntries: 50000
      queryTimeout: 2s
  exporters:
    - type: S3
      s3:
        endpoint: "https://minio.example:9000"
        bucket: netobserv-flows
        region: ""
        prefix: ""
        account: production-cluster   # cluster_id in object path
        batchSize: 5000
        writeTimeout: 60s
        format: Parquet               # wired to FLP as "parquet"
        compression: snappy           # CRD field; Parquet codec is FLP-internal
        credentials:
          name: netobserv-s3-creds   # keys: accessKeyId, secretAccessKey
        tls: {}
  consolePlugin:
    s3:
      enable: true                   # default: has S3 exporter && !loki.enable
```

## FLP pipeline write stage (flowBuffer)

Injected as a pipeline **write** stage (not a top-level JSON blob):

```yaml
parameters:
  - name: flowbuffer
    write:
      type: flowBuffer
      flowBuffer:
        maxEntries: 50000
        queryListenAddress: ":9200"
        queryTimeout: 2s
        serviceName: flowlogs-pipeline   # or flowlogs-pipeline-transformer (Kafka)
```

| API | Path |
|---|---|
| Cluster (fan-in) | `GET\|POST /api/flowbuffer/flows` |
| Peer (local) | `GET\|POST /api/flowbuffer/local/flows` |

Service / NetworkPolicy / RBAC use TCP **9200** (`query`). Peer discovery uses EndpointSlices (ClusterRole `netobserv-flp-peer-query`).

## FLP S3 encode stage

`parameters[]` encode type `s3` matches flowlogs-pipeline `EncodeS3`:

| Field | Source |
|---|---|
| `account`, `endpoint`, `bucket`, `prefix`, `format`, `batchSize`, `writeTimeout`, `secure` | CRD (`format` lowercased; `secure` from `https://` endpoint) |
| `accessKeyId` / `secretAccessKey` | **Plaintext** values loaded from the credentials Secret (FLP does not support file paths) |

Object layout (expected):

```text
s3://<bucket>/<prefix>/cluster_id=<account>/year=YYYY/month=MM/day=DD/hour=HH/part-….parquet
```

## Console plugin ConfigMap (`console-plugin-config` → `config.yaml`)

```yaml
flowBuffer:
  enable: true
  url: http://flowlogs-pipeline.<ns>.svc:9200   # or flowlogs-pipeline-transformer in Kafka mode
  timeout: 2s
s3:
  enable: true
  endpoint: …
  bucket: …
  region: …
  prefix: …
  account: …
  accessKeyPath: /var/s3-query-creds/accessKeyId
  secretKeyPath: /var/s3-query-creds/secretAccessKey
  skipTls: false
```

Frontend `features` may include `flowBuffer` and/or `s3`. Datasource dropdown: Loki, Prometheus, S3, Auto.

## Modes

- **Recommended:** Prom + S3 (Loki off, flowBuffer on, s3 on).
- **Buffer-only:** Loki off, no S3 — Console must warn (`RAW_FLOWS_BUFFER_ONLY`).
- **Loki + S3:** not recommended as dual stores (validation warning); S3 remains selectable explicitly in the Console while Auto prefers Loki.
