package httpie

import (
	"bufio"
	// "bytes"
	// "fmt"
	"io"
	"net/http"

	// "net/http/httputil"
	"os"
	// "io/ioutil"

	// "sync"
	// "syscall/js"

	"github.com/HexmosTech/httpie-go/exchange"
	"github.com/HexmosTech/httpie-go/flags"
	"github.com/HexmosTech/httpie-go/input"
	"github.com/HexmosTech/httpie-go/output"

	// "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Options struct {
	// Transport is applied to the underlying HTTP client. Use to mock or
	// intercept network traffic.  If nil, http.DefaultTransport will be cloned.
	Transport http.RoundTripper
}

type ExResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func Lama2Entry(cmdArgs []string, stdinBody io.Reader) interface{} {
	// Parse flags
	options := Options{}
	args, _, optionSet, _ := flags.Parse(cmdArgs)
	// if err != nil {
	// 	return ExResponse{}, err
	// }

	log.Info().Interface("commands from httie inside dependency", cmdArgs).Msg("cmdArgs")

	inputOptions := optionSet.InputOptions
	exchangeOptions := optionSet.ExchangeOptions
	exchangeOptions.Transport = options.Transport
	outputOptions := optionSet.OutputOptions

	// this shouldn't be hardcoded, but for testing
	// we are keeping it in this way
	// inputOptions.ReadStdin = true

	// Parse positional arguments
	in, _ := input.ParseArgs(args, stdinBody, &inputOptions)
	// if _, ok := errors.Cause(err).(*input.UsageError); ok {
	// 	usage.PrintUsage(os.Stderr)
	// 	return ExResponse{}, err
	// }
	// if err != nil {
	// 	return ExResponse{}, err
	// }

	// Send request and receive response
	status := Exchange(in, &exchangeOptions, &outputOptions)
	// if err != nil {
	// 	return ExResponse{}, err
	// }

	// if exchangeOptions.CheckStatus {
	// 	os.Exit(getExitStatus(status.StatusCode))
	// }

	return status
}

// func Main(options *Options) error {
// 	// Parse flags
// 	args, usage, optionSet, err := flags.Parse(os.Args)
// 	if err != nil {
// 		return err
// 	}
// 	inputOptions := optionSet.InputOptions
// 	exchangeOptions := optionSet.ExchangeOptions
// 	exchangeOptions.Transport = options.Transport
// 	outputOptions := optionSet.OutputOptions

// 	// this shouldn't be hardcoded, but for testing
// 	// we are keeping it in this way
// 	// inputOptions.ReadStdin = false

// 	// Parse positional arguments
// 	in, err := input.ParseArgs(args, os.Stdin, &inputOptions)
// 	if _, ok := errors.Cause(err).(*input.UsageError); ok {
// 		usage.PrintUsage(os.Stderr)
// 		return err
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	// Send request and receive response
// 	// status := Exchange(in, &exchangeOptions, &outputOptions)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// if exchangeOptions.CheckStatus {
// 	// 	os.Exit(getExitStatus(status.StatusCode))
// 	// }

// 	return nil
// }

func getExitStatus(statusCode int) int {
	if 300 <= statusCode && statusCode < 600 {
		return statusCode / 100
	}
	return 0
}

func Exchange(in *input.Input, exchangeOptions *exchange.Options, outputOptions *output.Options) interface{} {
	// Prepare printer
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// var bodyPlainBuffer bytes.Buffer
	// mWriter := io.MultiWriter(writer, &bodyPlainBuffer)

	// printer := output.NewPrinter(writer, outputOptions)
	// Build HTTP request
	request, _ := exchange.BuildHTTPRequest(in, exchangeOptions)
	// if err != nil {
	// 	return ExResponse{-1, "", map[string]string{}}, err
	// }

	log.Info().
		Str("Method", request.Method).
		Str("URL", request.URL.String()).
		Str("Host", request.Host).
		Msg("Received response body from API executor")
		
	responseMap := map[string]string{
		"Method": request.Method,
		"URL":    request.URL.String(),
		"Host":   request.Host,
	}
	return responseMap
	/*
		// Print HTTP request
		if outputOptions.PrintRequestHeader || outputOptions.PrintRequestBody {
			// `request` does not contain HTTP headers that HttpClient.Do adds.
			// We can get these headers by DumpRequestOut and ReadRequest.
			dump, err := httputil.DumpRequestOut(request, true)
			if err != nil {
				return ExResponse{-1, "", map[string]string{}}, err // should not happen
			}
			r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(dump)))
			if err != nil {
				return ExResponse{-1, "", map[string]string{}}, err // should not happen
			}
			defer r.Body.Close()

			// ReadRequest deletes Host header. We must restore it.
			if request.Host != "" {
				r.Header.Set("Host", request.Host)
			} else {
				r.Header.Set("Host", request.URL.Host)
			}

			if outputOptions.PrintRequestHeader {
				if err := printer.PrintRequestLine(r); err != nil {
					return ExResponse{-1, "", map[string]string{}}, err
				}
				if err := printer.PrintHeader(r.Header); err != nil {
					return ExResponse{-1, "", map[string]string{}}, err
				}
			}
			if outputOptions.PrintRequestBody {
				if err := printer.PrintBody(r.Body, r.Header.Get("Content-Type")); err != nil {
					return ExResponse{-1, "", map[string]string{}}, err
				}
			}
			fmt.Fprintln(writer)
			writer.Flush()
		}
	*/
	// func MakeRequest(url, method, body string, headers map[string]string) (<-chan Result, <-chan error) {
	// 	// Create channels
	// 	resultChan := make(chan Result, 1)
	// 	errorChan := make(chan error, 1)
	// 	// Start a goroutine to perform the HTTP request
	// 	go func() {
	// 		defer close(resultChan)
	// 		defer close(errorChan)
	// 		// Create an HTTP request
	// 		req, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewBufferString(body))
	// 		if err != nil {
	// 			errorChan <- err
	// 			return
	// 		}
	// 		// Set headers
	// 		for k, v := range headers {
	// 			req.Header.Set(k, v)
	// 		}
	// 		// Perform the request
	// 		client := &http.Client{}
	// 		resp, err := client.Do(req)
	// 		if err != nil {
	// 			errorChan <- err
	// 			return
	// 		}
	// 		defer resp.Body.Close()
	// 		// Read the response body
	// 		respBody, err := ioutil.ReadAll(resp.Body)
	// 		if err != nil {
	// 			errorChan <- err
	// 			return
	// 		}
	// 		// Send the result through the channel
	// 		resultChan <- Result{StatusCode: resp.StatusCode, Body: string(respBody), Headers: resp.Header}
	// 	}()
	// 	return resultChan, errorChan
	// // }

	// requestUrl := request.URL.String()

	// handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	resolve := args[0]
	// 	reject := args[1]

	// 	// Run this code asynchronously
	// 	go func() {
	// 		// Make the HTTP request
	// 		res, err := http.DefaultClient.Get(requestUrl)
	// 		if err != nil {
	// 			// Handle errors: reject the Promise if we have an error
	// 			errorConstructor := js.Global().Get("Error")
	// 			errorObject := errorConstructor.New(err.Error())
	// 			reject.Invoke(errorObject)
	// 			return
	// 		}
	// 		defer res.Body.Close()

	// 		// Read the response body
	// 		data, err := ioutil.ReadAll(res.Body)
	// 		if err != nil {
	// 			// Handle errors here too
	// 			errorConstructor := js.Global().Get("Error")
	// 			errorObject := errorConstructor.New(err.Error())
	// 			reject.Invoke(errorObject)
	// 			return
	// 		}

	// 		// "data" is a byte slice, so we need to convert it to a JS Uint8Array object
	// 		arrayConstructor := js.Global().Get("Uint8Array")
	// 		dataJS := arrayConstructor.New(len(data))
	// 		js.CopyBytesToJS(dataJS, data)

	// 		// Create a Response object and pass the data
	// 		responseConstructor := js.Global().Get("Response")
	// 		response := responseConstructor.New(dataJS)

	// 		// Resolve the Promise
	// 		resolve.Invoke(response)
	// 	}()

	// 	// The handler of a Promise doesn't return any value
	// 	return nil
	// })

	// promiseConstructor := js.Global().Get("Promise")
	// return promiseConstructor.New(handler)

	/*
		// Send HTTP request and receive HTTP request
		httpClient, err := exchange.BuildHTTPClient(exchangeOptions)
		if err != nil {
			return ExResponse{-1, "", map[string]string{}}, err
		}

		log.Info().
			Interface("Transport", httpClient.Transport).
			Msg("HttpClient details")


		resp, err := httpClient.Do(request)
		if err != nil {
			return ExResponse{-1, "", map[string]string{}}, errors.Wrap(err, "sending HTTP request")
		}
		defer resp.Body.Close()

		if outputOptions.PrintResponseHeader {
			if err := printer.PrintStatusLine(resp.Proto, resp.Status, resp.StatusCode); err != nil {
				return ExResponse{-1, "", map[string]string{}}, err
			}
			if err := printer.PrintHeader(resp.Header); err != nil {
				return ExResponse{-1, "", map[string]string{}}, err
			}
			writer.Flush()
		}

		if outputOptions.Download {
			file := output.NewFileWriter(in.URL, outputOptions)

			if err := printer.PrintDownload(resp.ContentLength, file.Filename()); err != nil {
				return ExResponse{-1, "", map[string]string{}}, err
			}
			writer.Flush()

			if err = file.Download(resp); err != nil {
				return ExResponse{-1, "", map[string]string{}}, err
			}
		} else {
			if outputOptions.PrintResponseBody {
				if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
					return ExResponse{-1, "", map[string]string{}}, err
				}
			}
		}

		respBody := ""
		switch printer.(type) {
		case *output.PrettyPrinter:
			pp := printer.(*output.PrettyPrinter)
			respBody = pp.BodyContent
		case *output.PlainPrinter:
			pp := printer.(*output.PlainPrinter)
			respBody = pp.BodyContent
		}

		headerMap := make(map[string]string, 0)
		for name, values := range resp.Header {
			for _, value := range values {
				headerMap[name] = value
			}
		}

		return ExResponse{resp.StatusCode, respBody, headerMap}, nil
		// promiseConstructor := js.Global().Get("Promise")
		// return promiseConstructor.New(handler)
	*/
}
