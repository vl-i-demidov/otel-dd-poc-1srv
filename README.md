# otel-datadog-poc

A POC Golang project that uses OpenTelemetry instrumentation for tracing and DataDog for profiling. All tracing and profiling data is sent to DataDog site via DataDog agent.
The purpose of this POC is to check if OTEL tracing and DD profiling can be correlated in the DataDog ('Code Hotspots' tab). 

## Pre-requisite

- `docker`
- `docker-compose`
- A valid DataDog API key

## Usage

### Start

`DD_API_KEY=<valid_api_key> make restart`

### Test
To issue an HTTP request and simulate high CPU load for 10 seconds: `make test SLEEP=10`

### Stop

`make stop`