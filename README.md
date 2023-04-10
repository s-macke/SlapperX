# SlapperX: A Simple Load Testing Tool

Slapper is a simple load testing tool written in Go, which can send HTTP requests to your server and display the results in a histogram. It provides basic performance metrics, such as request rate, response times, and error rates.

![slapper](img/example.gif)

## Features

- Basic performance metrics
- Histogram visualization of response time distribution
- Adjustable request rate, timeout, and worker count
- Supports multiple request targets
- Configurable ramp-up time for gradually increasing request rate


## Interface

![interface](img/interface.png)

## Installation

Just [download](https://github.com/s-macke/SlapperX/releases/tag/v0.2.3) a release or install SlapperX via

```bash
go install github.com/s-macke/slapperx
```

## Installation (Manual)

To use Slapper, you need to have Go installed on your machine. You can download Go from [the official website](https://golang.org/dl/).

After installing Go, clone this repository and build the binary:

```bash
git clone https://github.com/s-macke/SlapperX.git
cd SlapperX
go build
```

## Usage

Here is an example of how to use Slapper:

```bash
./slapper -targets targets.txt -workers 8 -timeout 30s -rate 50 -minY 0ms -maxY 100ms -rampup 10s
```

### Flags

- `-targets`: Targets file containing the list of URLs to be tested.
- `-workers`: Number of workers sending requests concurrently (default 8).
- `-timeout`: Request timeout duration (default 30 seconds).
- `-rate`: Desired request rate per second (default 50).
- `-minY`: Minimum Y-axis value for the histogram (default 0 milliseconds).
- `-maxY`: Maximum Y-axis value for the histogram (default 100 milliseconds).
- `-rampup`: Duration to ramp up to the desired request rate (default 0 seconds).
- 

### Keybindings

- `q`: Quit the program.
- `r`: Reset the statistics.
- `Ctrl+C`: Quit the program.

## Targets syntax

The targets file uses the same format as the Jetbrains `.http` files.
The full spec can be found [here](https://www.jetbrains.com/help/idea/exploring-http-syntax.html#short-form-for-get-requests)

	HTTP_METHOD url
	HEADERKEY1: HEADERVALUE1
	HEADERKEY2: HEADERVALUE2
	
	body
	
	###
	
	HTTP_METHOD url
	HEADERKEY1: HEADERVALUE1
	....
	
The http requests are separated by `###`.

## Contributing

Contributions are welcome! If you have found a bug or have a feature request, please create an issue or submit a pull request.

## License

This project is released under the [MIT License](https://opensource.org/licenses/MIT).
