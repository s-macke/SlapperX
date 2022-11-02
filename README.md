# slapper (enhanced)

__Simple load testing tool with real-time updated histogram of request timings__

![slapper](img/example.gif)

## Interface

![interface](img/interface.png)

## Usage
```bash
$ ./slapper -help
Usage of ./slapper:
  -maxY duration
    	max on Y axe (default 100ms)
  -minY duration
    	min on Y axe (default 0ms)
  -rate uint
    	Requests per second (default 50)
  -targets string
    	Targets file
  -timeout duration
    	Requests timeout (default 30s)
  -workers uint
    	Number of workers (default 8)

```

## Key bindings
* q, ctrl-c - quit
* r - reset stats

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
