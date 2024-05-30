# go-httping

httping is CLI tool like ping.  

## Install
### From source
```
go install github.com/ichisuke55/httping@latest
```

### Binary
release asset(TBD)

## Usage

```
httping -d URL [OTHER OPTIONS]
```

### Options

```
$ httping -h

Usage of httping:
  -X string
        HTTP method: GET, POST (default "GET")
  -c int
        number of times execute (default 5)
  -d string
        destination URL. e.g. 'http://localhost'
  -i float
        seconds between sending each httping request (default 1)
  -k bool
        skip SSL/TLS insecure verity (bool: default is false)
  -r bool
        disable redirect (bool: default is false)

```
