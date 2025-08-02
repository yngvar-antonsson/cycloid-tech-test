# Weather Aggregator Service

A simple HTTP server in Go that aggregates 5-day weather forecasts from multiple weather APIs.

## Configuration

The server reads its configuration from a YAML file you specify.

Example `config.yaml`:

```yaml
providers:
  openmeteo: {}
  weatherapi:
    APIKey: "your_weatherapi_key_here"
```
providers - map of provider names to their parameters.
For openmeteo, no parameters are needed.
For weatherapi, you must provide an APIKey.

## Command-line Flags

- `--config (string)` - path to directory containing config.yaml (default: `config.yml`)
- `--port (int)` - port number for HTTP server (default: `8080`)
- `--api-limit (int)` - timeout limit in seconds for API calls (default: `30`)

## Running the Application

Prepare your config.yaml as described above.
Build and run the application:
```go
go mod tidy
go build -o weather-app main.go
./weather-app --config=./config.yml --port=8080 --api-limit=30
```
The HTTP server will start and listen on the specified port.
Query the /weather endpoint with latitude and longitude query parameters, for example:
`http://localhost:8080/weather?lat=52.52&lon=13.41`
Expected response is a JSON aggregation of forecasts from configured providers.

## Running Tests

Run all tests for the project using:

```
go test -v ./...
```

## Adding providers

To add a provider, you need:
1) create a new code file in `/provider` directory.
2) write a `ProviderType` following the next API:
  - function, that returns name of the provider
    ```go
    func (o *OpenMeteo) Name() string
    ```
  - getter and setter (used to store some params in the structure, API_KEYs for example)
    ```go
    func (o *OpenMeteo) GetParams(string) any
    func (o *OpenMeteo) SetParams(string, any)
    ```
  - function that returns forecast (you can simplify your code using BuildForecastGetter function)
    ```go
    func (o *OpenMeteo) GetForecast(ctx context.Context, lat, lon string) (ForecastDay, error)
    ```
3) And finally add a new provider to the `handler/setup.go`.

## Project Structure

- `main.go` - application entry point, loads config and starts HTTP server
- `handler/` - contains HTTP handler and aggregator logic
- `provider/` - contains weather API providers implementations and request logic
- `config.yaml` - example configuration file for providers and parameters
