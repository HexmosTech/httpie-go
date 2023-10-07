package httpie

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"syscall/js"

	"github.com/HexmosTech/httpie-go/exchange"
	"github.com/HexmosTech/httpie-go/flags"
	"github.com/HexmosTech/httpie-go/input"
	"github.com/HexmosTech/httpie-go/output"
	"github.com/pkg/errors"
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

func Lama2Entry(cmdArgs []string, stdinBody io.Reader) (ExResponse, error) {
	// Parse flags
	// log.Info().Str("Req body", stdinBody).Msg("")
	log.Info().Interface("commands from httie inside dependency", cmdArgs).Msg("cmdArgs")
	options := Options{}
	args, usage, optionSet, err := flags.Parse(cmdArgs)
	if err != nil {
		log.Info().Err(err).Msg("Failed to parse flags")
		return ExResponse{}, err
	}
	inputOptions := optionSet.InputOptions
	exchangeOptions := optionSet.ExchangeOptions
	exchangeOptions.Transport = options.Transport
	outputOptions := optionSet.OutputOptions

	// this shouldn't be hardcoded, but for testing
	// we are keeping it in this way
	// inputOptions.ReadStdin = true

	// Parse positional arguments
	in, err := input.ParseArgs(args, stdinBody, &inputOptions)
	if _, ok := errors.Cause(err).(*input.UsageError); ok {
		usage.PrintUsage(os.Stderr)
		log.Warn().Err(err).Msg("Usage error while parsing positional arguments")
		return ExResponse{}, err
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse positional arguments")
		return ExResponse{}, err
	}

	// Send request and receive response
	status, err := Exchange(in, &exchangeOptions, &outputOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse positional arguments")
		return ExResponse{}, err
	}

	if exchangeOptions.CheckStatus {
		log.Error().Err(err).Msg("Error during Exchange")
		log.Info().Int("exitStatus", status.StatusCode).Msg("Checking status code for exit status")
		os.Exit(getExitStatus(status.StatusCode))
	}
	log.Info().Msg("Lama2Entry completed successfully")
	log.Info().Int("Req body", status.StatusCode).Msg("")
	return status, nil
}

func Main(options *Options) error {
	// Parse flags
	args, usage, optionSet, err := flags.Parse(os.Args)
	if err != nil {
		return err
	}
	inputOptions := optionSet.InputOptions
	exchangeOptions := optionSet.ExchangeOptions
	exchangeOptions.Transport = options.Transport
	outputOptions := optionSet.OutputOptions

	// this shouldn't be hardcoded, but for testing
	// we are keeping it in this way
	// inputOptions.ReadStdin = false

	// Parse positional arguments
	in, err := input.ParseArgs(args, os.Stdin, &inputOptions)
	if _, ok := errors.Cause(err).(*input.UsageError); ok {
		usage.PrintUsage(os.Stderr)
		return err
	}
	if err != nil {
		return err
	}

	// Send request and receive response
	status, err := Exchange(in, &exchangeOptions, &outputOptions)
	if err != nil {
		return err
	}

	if exchangeOptions.CheckStatus {
		os.Exit(getExitStatus(status.StatusCode))
	}

	return nil
}

func getExitStatus(statusCode int) int {
	if 300 <= statusCode && statusCode < 600 {
		return statusCode / 100
	}
	return 0
}

func ExchangeJS() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Extract arguments from JavaScript
		// Assuming args[0] is input, args[1] is exchangeOptions, and args[2] is outputOptions
		// You'll need to adapt this part to your needs

		in := &input.Input{}                   // Convert args[0] to this type
		exchangeOptions := &exchange.Options{} // Convert args[1] to this type
		outputOptions := &output.Options{}     // Convert args[2] to this type

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run this code asynchronously
			go func() {
				resp, err := Exchange(in, exchangeOptions, outputOptions)
				if err != nil {
					// Handle errors: reject the Promise if we have an error
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				// Convert the Go response to a JavaScript object
				jsResp := map[string]interface{}{
					"statusCode": resp.StatusCode,
					"body":       resp.Body,
					"headers":    resp.Headers,
				}

				// Resolve the Promise
				resolve.Invoke(jsResp)
			}()

			// The handler of a Promise doesn't return any value
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func Exchange(in *input.Input, exchangeOptions *exchange.Options, outputOptions *output.Options) (ExResponse, error) {
	// Prepare printer
	log.Info().Msg("Starting Exchange function")

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// var bodyPlainBuffer bytes.Buffer
	// mWriter := io.MultiWriter(writer, &bodyPlainBuffer)

	printer := output.NewPrinter(writer, outputOptions)
	// Build HTTP request
	request, err := exchange.BuildHTTPRequest(in, exchangeOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build HTTP request")
		return ExResponse{-1, "", map[string]string{}}, err
	}

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

	// Send HTTP request and receive HTTP request
	httpClient, err := exchange.BuildHTTPClient(exchangeOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build HTTP client")
		return ExResponse{-1, "", map[string]string{}}, err
	}

	// Create channels for response and error
	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)

	// Use a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		resp, err := httpClient.Do(request)
		if err != nil {
			errCh <- errors.Wrap(err, "sending HTTP request")
			return
		}
		respCh <- resp
	}()
	// Wait for the goroutine to finish
	wg.Wait()
	var resp *http.Response

	select {
	case err = <-errCh:
		log.Error().Err(err).Msg("Error sending HTTP request")
		return ExResponse{-1, "", map[string]string{}}, err
	case resp = <-respCh:
		defer resp.Body.Close()
	}

	// resp, err := httpClient.Do(request)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Error sending HTTP request")
	// 	return ExResponse{-1, "", map[string]string{}}, errors.Wrap(err, "sending HTTP request")
	// }
	// defer resp.Body.Close()

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
			log.Error().Err(err).Msg("Error printing download details")
			return ExResponse{-1, "", map[string]string{}}, err
		}
		writer.Flush()

		if err = file.Download(resp); err != nil {
			log.Error().Err(err).Msg("Error during file download")
			return ExResponse{-1, "", map[string]string{}}, err
		}
	} else {
		if outputOptions.PrintResponseBody {
			if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
				log.Error().Err(err).Msg("Error printing response body")
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
	log.Info().Msg("Exchange function completed successfully")

	return ExResponse{resp.StatusCode, respBody, headerMap}, nil
}
