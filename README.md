# RemoteExec

A simple REST API for sandboxed code execution using Gin, Docker and Redis. Submit code, execute it securely in containers, and track the execution status.

## Features

- **Sandboxed Execution** — Code runs inside Docker containers with resource limits.
- **Task Queue** — Jobs are stored in Redis and processed asynchronously.
- **REST API** — Submit code and poll for status using HTTP endpoints.
- **Easily scalable**: run any number of worker processes in parallel

## API Endpoints
| Methods | Endpoint          | Description                 |
| ------- | ----------------- | --------------------------- |
| POST    | `/execute`        | Submit code for execution   |
| GET     | `/status/:job_id` | Get status of submitted job |

## Examples

Execute 
```
curl -X POST http://localhost:8080/execute \
     -H "Content-Type: application/json" \
     -d '{"language": "python", "source_code": "print(\"Hello\")"}'

```

Status
```
curl http://localhost:8080/status/:job_id
```


## Execution Backend

- 🐳 **Docker** used to run untrusted code safely.
- 🧱 Uses Alpine images
- ⚙️ **Cgroups** and Docker resource flags used to limit memory, CPU, and time.
- 🔁 Workers pick up jobs from Redis, run containers, and store output.


## 📌 Usage

#### Running the API server

```
go run cmd/api-server/main.go
```

#### Running the Workers

```
go run cmd/worker/main.go
```

## Prerequisites

You'll need a Redis instance running locally. You can start one using Docker:
```
docker run -d --name redis-server -p 6379:6379 redis
```

Ensure the Python Docker image used by the worker (python:3.11-alpine) is available locally:
```
docker pull python:3.11-alpine
```


