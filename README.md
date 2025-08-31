# Document Scanning Rules Engine

This repository provides a minimal in-memory rules engine and HTTP service. It ingests uploaded documents (PDF, HTML, plain text, YAML, JSON, XML, or DOCX), normalizes them to text, evaluates that text against a configurable set of regex-based rules, and returns structured findings.

## Development

Run tests:

```bash
go test ./... -cover
```

## Deployment

For detailed deployment instructions, see the [Deployment Guide](DEPLOYMENT.md).

## API

### `POST /scan`
Upload a document to be scanned and receive a structured report of findings including rule descriptions.

**Request** – `multipart/form-data`

| Field | Type | Description |
|------|------|-------------|
| `file` | file | Document to scan. Supports `.pdf`, `.html`, `.txt`, `.yaml`, `.yml`, `.json`, `.xml`, `.docx` |

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

### `GET /health`
Health check endpoint.

**Response**
```json
{ "status": "ok" }
```

### `GET /docs`
Returns a JSON array of all available endpoints and their documentation.

**Response**

```json
[
  {
    "path": "/scan",
    "method": "POST",
    "description": "Upload a document to be scanned and receive a structured report of findings including rule descriptions.",
    "data_shapes": [
      {
        "name": "Request",
        "description": "multipart/form-data",
        "shape": "{\"file\": \"<file>\"}"
      },
      {
        "name": "Response",
        "description": "A structured report of findings.",
        "shape": "{\"file_id\":\"uploaded-filename\",\"findings\":[{\"rule_id\":\"rule-1\",\"severity\":\"high\",\"line\":3,\"context\":\"line containing match\",\"description\":\"rule description\"}]}"
      }
    ],
    "curl_example": "curl -X POST -F 'file=@/path/to/your/file.pdf' http://localhost:8080/scan"
  },
  {
    "path": "/rules/reload",
    "method": "POST",
    "description": "Replace the existing rules with a new set.",
    "data_shapes": [
      {
        "name": "Request",
        "description": "A JSON object containing the new rules.",
        "shape": "{\"rules\":[{\"id\":\"rule-1\",\"pattern\":\"secret\",\"severity\":\"high\"}]}"
      }
    ],
    "curl_example": "curl -X POST -H \"Content-Type: application/json\" -d '{\"rules\":[{\"id\":\"rule-1\",\"pattern\":\"secret\",\"severity\":\"high\"}]}' http://localhost:8080/rules/reload"
  },
  {
    "path": "/rules/load",
    "method": "POST",
    "description": "Load rules from a YAML file on disk.",
    "data_shapes": [
      {
        "name": "Request",
        "description": "A JSON object containing the path to the rules file.",
        "shape": "{\"path\":\"/etc/dws/rules.yaml\"}"
      }
    ],
    "curl_example": "curl -X POST -H \"Content-Type: application/json\" -d '{\"path\":\"/etc/dws/rules.yaml\"}' http://localhost:8080/rules/load"
  },
  {
    "path": "/health",
    "method": "GET",
    "description": "Health check endpoint.",
    "data_shapes": [
      {
        "name": "Response",
        "description": "A JSON object indicating the status of the service.",
        "shape": "{\"status\":\"ok\"}"
      }
    ],
    "curl_example": "curl http://localhost:8080/health"
  },
  {
    "path": "/docs",
    "method": "GET",
    "description": "Returns a JSON array of all available endpoints and their documentation.",
    "curl_example": "curl http://localhost:8080/docs"
  }
]
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