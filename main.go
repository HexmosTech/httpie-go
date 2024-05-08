package httpie

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"syscall/js"

	// "syscall/js"

	// "io"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	// "errors"
	// "fmt"
	"github.com/HexmosTech/httpie-go/exchange"
	"github.com/HexmosTech/httpie-go/flags"
	"github.com/HexmosTech/httpie-go/input"
	"github.com/HexmosTech/httpie-go/output"
	"github.com/pkg/errors"
	// "github.com/pkg/errors"
	// "syscall/js"
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

func getStringSlice(jsArray js.Value) []string {
	length := jsArray.Length()
	slice := make([]string, length)
	for i := 0; i < length; i++ {
		slice[i] = jsArray.Index(i).String()
	}
	return slice
}

func Int(i int) int {
	i, _ = strconv.Atoi(strconv.Itoa(i))
	return i
}

func Lama2Entry(cmdArgs []string, stdinBody io.Reader, proxyURL string, proxyUsername string, proxyPassword string, autoRedirect bool) (ExResponse, error) {
	fmt.Println("inisde Lama2 Entry File iteration number:0001")
	options := Options{}
	args, usage, optionSet, err := flags.Parse(cmdArgs)
	if err != nil {
		fmt.Println(err)
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
		return ExResponse{}, err
	}
	if err != nil {
		fmt.Println(err)
		return ExResponse{}, err
	}

	// Send request and receive response
	status, err := Exchange(in, &exchangeOptions, &outputOptions, proxyURL, proxyUsername, proxyPassword, autoRedirect)
	if err != nil {
		fmt.Println(err)
		return ExResponse{}, err
	}

	if exchangeOptions.CheckStatus {
		os.Exit(getExitStatus(status.StatusCode))
	}

	return status, nil
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
// 	status, err := Exchange(in, &exchangeOptions, &outputOptions, proxyURL,autoRedirect)
// 	if err != nil {
// 		return err
// 	}

// 	if exchangeOptions.CheckStatus {
// 		os.Exit(getExitStatus(status.StatusCode))
// 	}

// 	return nil
// }

func getExitStatus(statusCode int) int {
	if 300 <= statusCode && statusCode < 600 {
		return statusCode / 100
	}
	return 0
}

func Exchange(in *input.Input, exchangeOptions *exchange.Options, outputOptions *output.Options, proxyURL string, proxyUsername string, proxyPassword string, autoRedirect bool) (ExResponse, error) {
	// Prepare printer
	fmt.Println("inside exchange Function")
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	fmt.Println("Proxy params:", proxyURL)
	// var bodyPlainBuffer bytes.Buffer
	// mWriter := io.MultiWriter(writer, &bodyPlainBuffer)

	printer := output.NewPrinter(writer, outputOptions)
	// Build HTTP request
	request, err := exchange.BuildHTTPRequest(in, exchangeOptions)
	if err != nil {
		fmt.Println(err)
		return ExResponse{-1, "", map[string]string{}}, err
	}
	username := "proxyServer"
	password := "proxy22523146server"
	// auth := "proxyServer:proxy22523146server"

	auth := fmt.Sprintf("%s:%s", username, password)
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	request.Header.Add("proxyauth", basic)
	cookieValue := request.Header.Get("Cookie")
	if cookieValue != "" {
		request.Header.Add("CustomCookie", cookieValue)
	}

	request.RequestURI = ""
	// Print HTTP request
	if outputOptions.PrintRequestHeader || outputOptions.PrintRequestBody {
		// `request` does not contain HTTP headers that HttpClient.Do adds.
		// We can get these headers by DumpRequestOut and ReadRequest.
		dump, err := httputil.DumpRequestOut(request, true)
		if err != nil {
			fmt.Println(err)
			return ExResponse{-1, "", map[string]string{}}, err // should not happen
		}
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(dump)))
		if err != nil {
			fmt.Println(err)
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
				fmt.Println(err)
				return ExResponse{-1, "", map[string]string{}}, err
			}
			if err := printer.PrintHeader(r.Header); err != nil {
				fmt.Println(err)
				return ExResponse{-1, "", map[string]string{}}, err
			}
		}
		if outputOptions.PrintRequestBody {
			if err := printer.PrintBody(r.Body, r.Header.Get("Content-Type")); err != nil {
				fmt.Println(err)
				return ExResponse{-1, "", map[string]string{}}, err
			}
		}
		fmt.Fprintln(writer)
		writer.Flush()
	}

	// Send HTTP request and receive HTTP request
	httpClient, err := exchange.BuildHTTPClient(exchangeOptions, proxyURL, proxyUsername, proxyPassword, autoRedirect, request)
	fmt.Println("after build_http_Client Function")
	if err != nil {
		fmt.Println(err)
		return ExResponse{-1, "", map[string]string{}}, err
	}
	fmt.Println("Making HTTP request", request)
	resp, err := httpClient.Do(request)
	// resp.Header.Set("Access-Control-Allow-Origin", "*")
	// resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, User-Agent")

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if outputOptions.PrintResponseHeader {
		if err := printer.PrintStatusLine(resp.Proto, resp.Status, resp.StatusCode); err != nil {
			fmt.Println(err)
			// return ExResponse{-1, "", map[string]string{}}, err
		}
		if err := printer.PrintHeader(resp.Header); err != nil {
			// return ExResponse{-1, "", map[string]string{}}, err
			fmt.Println(err)
		}
		writer.Flush()
	}

	if outputOptions.Download {
		file := output.NewFileWriter(in.URL, outputOptions)

		if err := printer.PrintDownload(resp.ContentLength, file.Filename()); err != nil {
			// return ExResponse{-1, "", map[string]string{}}, err
			fmt.Println(err)
		}
		writer.Flush()

		if err = file.Download(resp); err != nil {
			// return ExResponse{-1, "", map[string]string{}}, err
			fmt.Println(err)
		}
	} else {
		if outputOptions.PrintResponseBody {
			if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
				// return ExResponse{-1, "", map[string]string{}}, err
				fmt.Println(err)
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
}
