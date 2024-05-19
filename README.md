# httpie-go

[![CircleCI](https://circleci.com/gh/nojima/httpie-go.svg?style=shield)](https://circleci.com/gh/nojima/httpie-go)

![httpie-go screenshot](./docs/images/screenshot.png)

**httpie-go** (`ht`) is a user-friendly HTTP client CLI.
Requests can be issued with fewer types compared to `curl`.
Responses are displayed with syntax highlighting.

httpie-go is a clone of [httpie](https://httpie.org/).
Since httpie-go is written in Go, it is a single binary and does not require a heavy runtime.

## Examples

This example sends a GET request to http://httpbin.org/get.

```bash
$ ht GET httpbin.org/get
```

The second example sends a POST request with JSON body `{"hello": "world", "foo": "bar"}`.

```bash
$ ht POST httpbin.org/post hello=world foo=bar
```

You can see the request that is being sent with `-v` option.

```bash
$ ht -v POST httpbin.org/post hello=world foo=bar
```

Request HTTP headers can be specified in the form of `key:value`.

```bash
$ ht -v POST httpbin.org/post X-Foo:foobar
```

Disable TLS verification.

```bash
$ ht --verify=no https://httpbin.org/get
```

Download a file.

```bash
$ ht --download <any url you want>
```

## Documents

Although httpie-go does not currently have documents, you can refer to the original [httpie's documentation](https://httpie.org/doc) since httpie-go is a clone of httpie.
Note that some minor options are yet to be implemented in httpie-go.

## How to build

```
make
```

For non-standard Linux system like Android [termux](https://termux.com/), use following method to avoid the DNS issue.

```
make build-termux
```


# How build tags work.

We use build tags or build constraints to separate build process for different platforms. Build tags are found at the top of the file as a comment. The file will be included only if the tag is present in the build command.
eg: 
```go
//go:build cli
```
#### build command using wasm tag
```sh
GOOS=js GOARCH=wasm go build -tags=wasm  -o static/main.wasm
```
#### build command using cli tag

```sh
go build -tags=cli  -o static/main.wasm
```

Note : The first line of the file should be followed by an empty line to make it a valid build tag statement.

If there is an `!` mark then it means the `not` operation on the build tags.

```go
//go:build !windows
```
This will exclude the file if a `windows` build tag is used.

Here we have `wasm` and `cli` build tags to switch between wasm and cli builds.