# Document Scanning Rules Engine

This repository provides a minimal in-memory rules engine and HTTP service. It ingests uploaded documents (PDF, HTML, plain text, or YAML), normalizes them to text, evaluates that text against a configurable set of regex-based rules, and returns structured findings.

## Development

Run tests:

```bash
go test ./... -cover
```

## API

### `POST /scan`
Upload a document to be scanned.

**Request** – `multipart/form-data`

| Field | Type | Description |
|------|------|-------------|
| `file` | file | Document to scan. Supports `.pdf`, `.html`, `.txt`, `.yaml`, `.yml` |

**Response** – array of findings

```json
[
  {
    "file_id": "uploaded-filename",
    "rule_id": "rule-1",
    "severity": "high",
    "line": 3,
    "context": "line containing match",
    "description": "rule description"
  }
]
```

### `POST /rules/reload`
Replace the existing rules with a new set.

**Request**
```json
{
  "rules": [
    { "id": "rule-1", "pattern": "secret", "severity": "high" }
  ]
}
```

**Response**
- `200 OK` on success

### `POST /rules/load`
Load rules from a YAML file on disk.

**Request**
```json
{ "path": "/etc/dws/rules.yaml" }
```

**Response**
- `200 OK` on success

### `POST /report`
Upload a document and receive a structured report of findings including rule descriptions.

**Request** – `multipart/form-data`

| Field | Type | Description |
|------|------|-------------|
| `file` | file | Document to scan |

**Response**

```json
{
  "file_id": "uploaded-filename",
  "findings": [
    {
      "rule_id": "rule-1",
      "severity": "high",
      "line": 3,
      "context": "line containing match",
      "description": "rule description"
    }
  ]
}
```

### `GET /health`
Health check endpoint.

**Response**
```json
{ "status": "ok" }
```

## Data Structures

### Rule
```json
{
  "id": "unique rule identifier",
  "pattern": "regex pattern",
  "severity": "severity string",
  "description": "rule description"
}
```

### Finding
```json
{
  "file_id": "id of scanned file",
  "rule_id": "matched rule id",
  "severity": "severity string",
  "line": 1,
  "context": "matching line snippet",
  "description": "rule description"
}
```

## Kubernetes Deployment

Each application can run its own service instance with an isolated rule set. Package rule files into ConfigMaps and mount them at `/etc/dws/rules.yaml`. Set the `RULES_FILE` environment variable so the service loads the desired rules at startup.

The service fails to start if the referenced rules file is missing or invalid, ensuring that each pod only runs with an explicit configuration.

Example ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rules-a
data:
  rules.yaml: |
    rules:
    - id: r1
      pattern: secret
      severity: high
```

Pod snippet:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: scanner-a
spec:
  containers:
  - name: dws
    image: dws:latest
    env:
    - name: RULES_FILE
      value: /etc/dws/rules.yaml
    volumeMounts:
    - name: rules
      mountPath: /etc/dws
  volumes:
  - name: rules
    configMap:
      name: rules-a
```

Deploy a separate pod with a different ConfigMap for each rule set.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.