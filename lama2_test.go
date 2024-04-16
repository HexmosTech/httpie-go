package httpie

import (
	"fmt"
	"strings"
	"testing"
)

func TestLama2Entry(t *testing.T) {
	// Example command line arguments
	cmdArgs := []string{"ht", "get", "https://google.com"}
	// cmdArgs := []string{"GET", "https://httpbin.org/get"}

	// Example stdin input
	stdinBody := strings.NewReader("")

	// Example proxy parameters
	proxyURL := "http://proxyserver.hexmos.com:8080/"
	proxyUsername := "proxyServer"
	proxyPassword := "proxy22523146server"

	// Example auto redirect option
	autoRedirect := true

	// Call Lama2Entry function
	response, err := Lama2Entry(cmdArgs, stdinBody, proxyURL, proxyUsername, proxyPassword, autoRedirect)

	// Check if there was an error
	if err != nil {
		t.Errorf("Error executing Lama2Entry: %v", err)
	}

	// Print response information
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Response Body: %s\n", response.Body)
	fmt.Println("Headers:")
	for key, value := range response.Headers {
		fmt.Printf("%s: %s\n", key, value)
	}
}
