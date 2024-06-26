package output

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
)

type PlainPrinter struct {
	writer      io.Writer
	BodyContent string
}

func NewPlainPrinter(writer io.Writer) Printer {
	return &PlainPrinter{
		writer: writer,
	}
}

func (p *PlainPrinter) GetPlainPrinter() Printer {
	return p
}

func (p *PlainPrinter) PrintStatusLine(proto string, status string, statusCode int) error {
	fmt.Fprintf(p.writer, "%s %s\n", proto, status)
	return nil
}

func (p *PlainPrinter) PrintRequestLine(req *http.Request) error {
	fmt.Fprintf(p.writer, "%s %s %s\n", req.Method, req.URL, req.Proto)
	return nil
}

func (p *PlainPrinter) PrintHeader(header http.Header) error {
	for name, values := range header {
		for _, value := range values {
			fmt.Fprintf(p.writer, "%s: %s\n", name, value)
		}
	}
	fmt.Fprintln(p.writer)
	return nil
}

func (p *PlainPrinter) PrintBody(body io.Reader, contentType string) error {
	s, _ := io.ReadAll(body)
	p.BodyContent = string(s)
	_, err := io.Copy(p.writer, strings.NewReader(p.BodyContent))
	if err != nil {
		return errors.Wrap(err, "printing body")
	}
	return nil
}

func (p *PlainPrinter) PrintDownload(length int64, filename string) error {
	fmt.Fprintf(p.writer, "Downloading %sB to \"%s\"\n", bytefmt.ByteSize(uint64(length)), filename)
	return nil
}
