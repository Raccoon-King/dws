# Curl Examples for Test Files

This document provides curl example commands for testing the DWS API with each test file in the `testfiles/` directory. The server is assumed to be running locally on port 8080.

## Endpoint: /scan

Upload a file to be scanned using the default ruleset.

### sample.html
```bash
curl -X POST -F "file=@testfiles/sample.html" http://localhost:8080/scan
```

### sample.htm
```bash
curl -X POST -F "file=@testfiles/sample.htm" http://localhost:8080/scan
```

### sample.json
```bash
curl -X POST -F "file=@testfiles/sample.json" http://localhost:8080/scan
```

### sample.pdf
```bash
curl -X POST -F "file=@testfiles/sample.pdf" http://localhost:8080/scan
```

### sample.txt
```bash
curl -X POST -F "file=@testfiles/sample.txt" http://localhost:8080/scan
```

### sample.xml
```bash
curl -X POST -F "file=@testfiles/sample.xml" http://localhost:8080/scan
```

### sample.yaml
```bash
curl -X POST -F "file=@testfiles/sample.yaml" http://localhost:8080/scan
```

### sample.yml
```bash
curl -X POST -F "file=@testfiles/sample.yml" http://localhost:8080/scan
```

### rules_json_test.yaml
```bash
curl -X POST -F "file=@testfiles/rules_json_test.yaml" http://localhost:8080/scan
```

### rules_json_test2.yaml
```bash
curl -X POST -F "file=@testfiles/rules_json_test2.yaml" http://localhost:8080/scan
```

## Endpoint: /ruleset

Upload a file to be scanned against a specific ruleset (requires a `rules/{rule}.yaml` file).

### With sample.html
```bash
curl -X POST -F "file=@testfiles/sample.html" "http://localhost:8080/ruleset?rule=testrules"
```

*Note: Replace "testrules" with the name of your ruleset file (without .yaml extension). The ruleset file should exist as "rules/testrules.yaml".*

### Using test ruleset files

Assuming you have placed the test rules files in the "rules/" directory as "rules/rules_json_test.yaml" and "rules/rules_json_test2.yaml":

#### Using rules_json_test.yaml
```bash
curl -X POST -F "file=@testfiles/sample.html" "http://localhost:8080/ruleset?rule=rules_json_test"
```

#### Using rules_json_test2.yaml
```bash
curl -X POST -F "file=@testfiles/sample.txt" "http://localhost:8080/ruleset?rule=rules_json_test2"
```

## Other Endpoints

### View API Documentation
```bash
curl http://localhost:8080/docs
```

### Health Check
```bash
curl http://localhost:8080/health
```

### Reload Rules (JSON)
```bash
curl -X POST -H "Content-Type: application/json" -d "{\"rules\":[{\"id\":\"rule-1\",\"pattern\":\"secret\",\"severity\":\"high\"}]}" http://localhost:8080/rules/reload
```

### Load Rules from File
```bash
curl -X POST -H "Content-Type: application/json" -d "{\"path\":\"default.yaml\"}" http://localhost:8080/rules/load
```

## Notes
- Replace `http://localhost:8080` with your actual server URL if different.
- Ensure the rules files are accessible to the server (e.g., in the correct directory).
- Ensure the DWS server is running before executing these commands (e.g., `go run main.go`).
- For large files or slow networks, consider adding `--connect-timeout 30` to avoid timeouts.
- Commands are formatted for Windows Command Prompt; single quotes are replaced with double quotes for compatibility.
