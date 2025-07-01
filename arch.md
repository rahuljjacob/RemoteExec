Should serve as a REST API 

Code submission part
- User can submit CODE using a POST request
- User can submit polling requests using GET
- Validate Code 
- Enqueue into a task queue (redis)
- Make task ID


Task Queue
- Store Tasks in Valkey
- Use valkey REST API
- Pass Tasks from here to worker processes

Sandboxed Execution
- Run code in docker
- Limit Resources for the image (cgroups???)
- Use some base image (alpine + gcc or smth)


Workflow

- User Submits code using POST request -> 
  REST API Handles getting code -> 
  Validate Code (LOC etc...) -> 
  Put Code in Task Queue (REDIS) ->
  Workers execute code and return val (Go files ig??)
  

Endpoints:
/execute (POST)
/status/:job_id (GET)
/logs/:job_id (GET) (optional)


Structure
rce-api/
├── cmd/
│   ├── api-server/         # main.go (imports internal packages)
│   └── worker/             # main.go (imports internal packages)
│
├── internal/
│   ├── models/             # package models
│   └── utils/              # package utils
│
├── config/                 # package config
│
├── go.mod
└── go.sum


Curls For Testing

execute endpoint : 
curl -X POST http://localhost:8080/execute \
     -H "Content-Type: application/json" \
     -d '{"language": "python", "source_code": "print(\"Hello\")"}'

