# HTTP Traffic Generator

This is a Golang-based HTTP client application designed to perform multiple HTTP requests concurrently with detailed logging and statistics. The client supports various HTTP methods, custom headers, and request bodies, making it suitable for testing and benchmarking web servers.

## Features

- Support for multiple HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE)
- Custom headers and request bodies
- Configurable request timeout and interval between requests
- Detailed logging of errors
- JSON output of request results
- Statistics on total, failed requests, and average response time
- Concurrency control with a maximum number of concurrent requests

## Installation

1. Ensure you have Golang installed on your system.
2. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/http-traffic-generator.git
   cd http-traffic-generator
   ```

3. Build the application:

   ```bash
   go build -o http_generator
   ```

## Usage

Run the application with the desired flags:

```sh
./http_generator -url "http://example.com" -n 10 -method "GET" -timeout 10 -headers "Content-Type=application/json" -interval 1000 -output "results.json" -errorlog "errors.log" -maxconcurrent 5
```

### Flags

- `-url`: URL to send requests to (default: "http://example.com")
- `-n`: Number of requests (default: 10)
- `-method`: HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE) (default: "GET")
- `-body`: Request body (for POST, PUT, PATCH method) (default: "")
- `-timeout`: Request timeout in seconds (default: 10)
- `-headers`: Custom headers (format: key1=value1,key2=value2) (default: "")
- `-interval`: Interval between requests in milliseconds (default: 0)
- `-output`: Output file to save results (default: "results.json")
- `-errorlog`: File to log errors (default: "errors.log")
- `-maxconcurrent`: Maximum number of concurrent requests (default: 5)

### Examples

#### GET Request

```sh
./http_generator -url "http://example.com" -n 10 -method "GET" -timeout 10 -output "results.json" -errorlog "errors.log"
```

#### POST Request with Body

```sh
./http_generator -url "http://example.com" -n 10 -method "POST" -body '{"key":"value"}' -headers "Content-Type=application/json" -timeout 10 -output "results.json" -errorlog "errors.log"
```

#### Limiting Concurrent Requests

```sh
./http_generator -url "http://example.com" -n 10 -method "GET" -timeout 10 -maxconcurrent 3 -output "results.json" -errorlog "errors.log"
```

## Output

The application produces two output files:

1. **Results File (JSON)**: Contains the results of each request, including status, duration, response length, and any errors.
2. **Error Log File**: Contains detailed error messages for any requests that failed.

### Results File (results.json)

The results file is a JSON array where each object represents the result of a single request. Here is an example:

```json
[
    {
        "status": "200 OK",
        "duration": "150ms",
        "response_length": 1234
    },
    {
        "status": "500 Internal Server Error",
        "duration": "200ms",
        "response_length": 567,
        "error": "Internal Server Error"
    }
]
```

### Error Log File (errors.log)

The error log file contains detailed error messages, each prefixed with a timestamp. Here is an example:

```
2024-07-29T15:30:45Z: Error creating request: some error message
2024-07-29T15:30:46Z: Error: Internal Server Error
2024-07-29T15:30:47Z: Error reading response body: some error message
```

## Statistics

At the end of execution, the application prints the following statistics:

- Total requests
- Failed requests
- Average response time

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
