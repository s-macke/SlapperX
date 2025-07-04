# SlapperX: A Simple Load Testing Tool

SlapperX is a simple load testing tool written in Go, which can send HTTP requests to your server and display the results in a histogram. It provides basic performance metrics, such as request rate, response times, and error rates.

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
go install github.com/s-macke/SlapperX@latest
```

## Installation (Manual)

To use SlapperX, you need to have Go installed on your machine. You can download Go from [the official website](https://golang.org/dl/).

After installing Go, clone this repository and build the binary:

```bash
git clone https://github.com/s-macke/SlapperX.git
cd SlapperX
go build
```

## Usage

Here is an example of how to use SlapperX:

```bash
./slapperx -targets targets.http -workers 8 -timeout 30s -rate 50 -minY 0ms -maxY 100ms -rampup 10s
```

### Flags

- `-targets`: Targets file containing the REST request data to be tested in the [.http format](https://www.jetbrains.com/help/idea/exploring-http-syntax.html).
- `-workers`: Number of workers sending requests concurrently (default 50).
- `-timeout`: Request timeout duration (default 30 seconds).
- `-rate`: Desired request rate per second (default 50).
- `-minY`: Minimum Y-axis value for the histogram (default 0 milliseconds).
- `-maxY`: Maximum Y-axis value for the histogram (default 100 milliseconds).
- `-rampup`: Duration to ramp up to the desired request rate (default 0 seconds).

### Keybindings

- `q`: Quit the program.
- `r`: Reset the statistics.
- `k`: Increase request rate by 10
- `j`: Decrease request rate by 10
- `Ctrl+C`: Quit the program.

## Targets syntax

The targets file follows the same format as the JetBrains `.http` files.
You can find the full specification in the
[JetBrains documentation](https://www.jetbrains.com/help/idea/exploring-http-syntax.html).

Here is an example:

```
GET https://api.example.com/users
Authorization: Bearer your_token_here
Content-Type: application/json

###

POST https://api.example.com/users
Authorization: Bearer your_token_here
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john.doe@example.com"
}

###

PUT https://api.example.com/users/123
Authorization: Bearer your_token_here
Content-Type: application/json

{
  "email": "updated.email@example.com"
}
```

In this example, we have three HTTP requests:

1. A `GET` request to retrieve a list of users.
2. A `POST` request to create a new user.
3. A `PUT` request to update an existing user's email.

Each request includes headers, such as `Authorization` and `Content-Type`, and uses the `###` separator to distinguish between requests.


## Contributing

Contributions are welcome! If you have found a bug or have a feature request, please create an issue or submit a pull request.

## License

This project is released under the [MIT License](https://opensource.org/licenses/MIT).
