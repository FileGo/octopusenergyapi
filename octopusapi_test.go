package octopusenergyapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		output := checkPostcode(test.input)
		assert.Equal(t, test.expected, output)
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

		if test.errExpected {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.expected, out)
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
		{
			"ftp://not a real url",
			"testapikey",
			"",
			true,
		},
	}

	for _, test := range tests {
		c, err := NewClient(test.baseURL, test.APIkey, http.DefaultClient)
		if test.errExpected {
			assert.NotNil(t, err)
		} else {
			if assert.Nil(t, err) {
				assert.IsType(t, client{}, c)
				assert.Equal(t, test.baseURLExpected, c.baseURL)
			}
		}
	}

}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return client, s.Close
}

func TestListProducts(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		f, err := os.Open("./testdata/listproducts.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		products, err := client.ListProducts()
		if assert.Nil(t, err) {
			assert.Len(t, products, 100)
		}
	})

	t.Run("http_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.ListProducts()
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http error")
			}
		}
	})

	t.Run("json_error", func(t *testing.T) {
		f, err := os.Open("./testdata/error.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.ListProducts()
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unmarshal json")
		}
	})

	t.Run("nilURL_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.ListProducts()
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "http get error")
		}
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		f, err := os.Open("./testdata/getproduct.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		expProduct := Product{
			Code:    "VAR-17-01-11",
			IsGreen: false,
		}
		product, err := client.GetProduct("VAR-17-01-11")

		if assert.Nil(t, err) {
			// Only check the values we've set
			assert.Equal(t, expProduct.Code, product.Code)
			assert.Equal(t, expProduct.IsGreen, product.IsGreen)

			assert.Len(t, product.SingleRegisterElecTariffs, 14)
			assert.Len(t, product.DualRegisterElecTariffs, 14)
		}
	})

	t.Run("http_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetProduct("productcode")
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http error")
			}
		}
	})

	t.Run("json_error", func(t *testing.T) {
		f, err := os.Open("./testdata/error.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetProduct("productcode")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unmarshal json")
		}
	})

	t.Run("nilURL_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetProduct("productcode")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "http get error")
		}
	})
}

func TestGetMeterPoint(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		f, err := os.Open("./testdata/getmeterpoint.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		expMP := MeterPoint{
			MPAN:         "0123456789",
			ProfileClass: 1,
			GSP: GridSupplyPoint{
				GSPGroupID: "_A",
			},
		}
		mp, err := client.GetMeterPoint(expMP.MPAN)

		if assert.Nil(t, err) {
			// TODO
			assert.Equal(t, expMP.MPAN, mp.MPAN)
			assert.Equal(t, expMP.ProfileClass, mp.ProfileClass)
			assert.Equal(t, expMP.GSP.GSPGroupID, mp.GSP.GSPGroupID)
		}
	})

	t.Run("no_gsp_error", func(t *testing.T) {
		f, err := os.Open("./testdata/getmeterpoint_nogsp.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetMeterPoint("0123456789")

		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "no grid supply point found")
		}
	})

	t.Run("http_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetMeterPoint("0123456789")
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http error")
			}
		}
	})

	t.Run("json_error", func(t *testing.T) {
		f, err := os.Open("./testdata/error.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetMeterPoint("0123456789")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unmarshal json")
		}
	})

	t.Run("nilURL_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetMeterPoint("0123456789")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "http get error")
		}
	})
}

func TestGetGridSupplyPoint(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		f, err := os.Open("./testdata/getgridsupplypoint.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		gsp, err := client.GetGridSupplyPoint("SW1A 1AA")

		if assert.Nil(t, err) {
			assert.Equal(t, GSPs[0], gsp) // GSPs[0] represents "_A"
		}
	})

	t.Run("many_gsp_error", func(t *testing.T) {
		f, err := os.Open("./testdata/getgridsupplypoint_err.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetGridSupplyPoint("SW1A 1AA")

		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "more than one")
		}
	})

	t.Run("no_gsp_error", func(t *testing.T) {
		f, err := os.Open("./testdata/getgridsupplypoint_nogsp.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetGridSupplyPoint("SW1A 1AA")

		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unknown grid supply point")
		}
	})

	t.Run("postcode_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetGridSupplyPoint("invalid_postcode")
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "invalid postcode")
			}
		}
	})

	t.Run("http_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetGridSupplyPoint("SW1A 1AA")
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http error")
			}
		}
	})

	t.Run("json_error", func(t *testing.T) {
		f, err := os.Open("./testdata/error.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetGridSupplyPoint("SW1A 1AA")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unmarshal json")
		}
	})

	t.Run("nilURL_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("", "fakeapikey", httpClient)
		assert.Nil(t, err)

		_, err = client.GetGridSupplyPoint("SW1A 1AA")
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "http get error")
		}
	})
}

func TestGetMeterConsumption(t *testing.T) {
	mpan := "0123456789"
	serialNo := "0123456789"

	t.Run("options", func(t *testing.T) {
		timeFrom, err := time.Parse("2006-01-02 15:04:05", "2020-01-02 12:23:34")
		assert.Nil(t, err)
		timeTo, err := time.Parse("2006-01-02 15:04:05", "2020-01-03 12:23:34")
		assert.Nil(t, err)
		pageSize := 10

		options := ConsumptionOption{
			From:     timeFrom,
			To:       timeTo,
			OrderBy:  "asc",
			GroupBy:  "hour",
			PageSize: pageSize,
		}

		f, err := os.Open("./testdata/consumption.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			assert.Equal(t, options.From.Format(iso8601), q.Get("period_from"))
			assert.Equal(t, options.To.Format(iso8601), q.Get("period_to"))
			assert.Equal(t, fmt.Sprint(pageSize), q.Get("page_size"))
			assert.Equal(t, options.OrderBy, q.Get("order_by"))
			assert.Equal(t, options.GroupBy, q.Get("group_by"))

			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetMeterConsumption(mpan, serialNo, options)
			assert.Nil(t, err)
		}
	})

	t.Run("json_error", func(t *testing.T) {
		f, err := os.Open("./testdata/error.json")
		assert.Nil(t, err)
		defer f.Close()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, f)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetMeterConsumption(mpan, serialNo, ConsumptionOption{})
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "unmarshal json")
			}
		}
	})

	t.Run("http_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("https://testapi.testdomain/", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetMeterConsumption(mpan, serialNo, ConsumptionOption{})
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http error")
			}
		}
	})

	t.Run("nilURL_error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		httpClient, teardown := testingHTTPClient(h)
		defer teardown()

		client, err := NewClient("", "fakeapikey", httpClient)
		if assert.Nil(t, err) {
			_, err = client.GetMeterConsumption(mpan, serialNo, ConsumptionOption{})
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), "http get error")
			}
		}
	})
}
