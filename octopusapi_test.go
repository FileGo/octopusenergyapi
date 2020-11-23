package octopusenergyapi

import (
	"net/http"
	"reflect"
	"testing"
)

func TestCheckPostcode(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"SW1A 1AA", true},
		{"sW1A 1aA", true},
		{"E20 2ST", true},
		{"e20 2st", true},
		{"this is not a postcode", false},
	}

	for _, test := range tests {
		if output := checkPostcode(test.input); output != test.expected {
			t.Errorf("input: %s, out: %t, expected: %t", test.input, output, test.expected)
		}
	}
}

func TestUrlAddUsername(t *testing.T) {
	tests := []struct {
		URL         string
		username    string
		expected    string
		errExpected bool
	}{
		{"http://www.google.com/", "user", "http://user:@www.google.com/", false},
		{"https://www.google.com/", "user", "https://user:@www.google.com/", false},
		{"10928301####$$$%%", "user", "", true},
	}

	for _, test := range tests {
		out, err := urlAddUsername(test.URL, test.username)
		if err != nil && !test.errExpected {
			t.Errorf("url: %s, username: %s, error not expected, error returned: %v", test.URL, test.username, err)
		}

		if out != test.expected {
			t.Errorf("url: %s, username: %s, expected: %s, output: %s", test.URL, test.username, test.expected, out)
		}
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		baseURL         string
		APIkey          string
		baseURLExpected string
		errExpected     bool
	}{
		{
			"https://localhost/",
			"testapikey",
			"https://testapikey:@localhost",
			false,
		},
		{
			"https://localhost//",
			"testapikey",
			"https://testapikey:@localhost",
			false,
		},
	}

	for _, test := range tests {
		c, err := NewClient(test.baseURL, test.APIkey, *http.DefaultClient)
		if test.errExpected && err == nil {
			t.Errorf("baseURL: %v, API key: %v, error expected but not returned", test.baseURL, test.APIkey)
		}

		if !test.errExpected && err != nil {
			t.Errorf("baseURL: %v, API key: %v, unexpected error returned: %v", test.baseURL, test.APIkey, err)
		}

		if reflect.DeepEqual(c, client{}) {
			t.Errorf("error not expected: %v", err)
		}

		if c.baseURL != test.baseURLExpected {
			t.Errorf("baseURL: %v, APIkey: %v, output: %v, expected: %v", test.baseURL, test.APIkey, c.baseURL, test.baseURLExpected)
		}
	}

}
