package octopusenergyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var postcodeRegex *regexp.Regexp

func init() {
	// Compile postcode regexp
	postcodeRegex = regexp.MustCompile(`^([Gg][Ii][Rr] 0[Aa]{2})|((([A-Za-z][0-9]{1,2})|(([A-Za-z][A-Ha-hJ-Yj-y][0-9]{1,2})|(([AZa-z][0-9][A-Za-z])|([A-Za-z][A-Ha-hJ-Yj-y][0-9]?[A-Za-z])))) [0-9][A-Za-z]{2})$`)
}

// NewClient returns a client
func NewClient(APIkey string, httpClient *http.Client) (*Client, error) {
	// Empty APIkey is not permitted
	APIkey = strings.TrimSpace(APIkey)
	if len(APIkey) == 0 {
		return nil, errors.New("API key should not be empty")
	}

	// Add APIkey as username to base URL
	baseURL, err := urlAddUsername(baseURL, APIkey)
	if err != nil {
		return nil, errors.Errorf("unable to add username to url: %v", err)
	}

	return &Client{
		URL:        baseURL,
		httpClient: httpClient,
	}, nil
}

// GetMeterPoint retrieves a meter point for a given MPAN
// https://developer.octopus.energy/docs/api/#electricity-meter-points
func (c *Client) GetMeterPoint(mpan string) (MeterPoint, error) {
	data := struct {
		GspID        string `json:"gsp"`
		MPAN         string `json:"mpan"`
		ProfileClass int    `json:"profile_class"`
	}{}

	err := c.do(fmt.Sprintf("electricity-meter-points/%s/", mpan), &data)
	if err != nil {
		return MeterPoint{}, errors.Errorf("error retrieving meterpoint: %v", err)
	}

	// Mask JSON struct into MeterPoint
	mPoint := MeterPoint{
		MPAN:         data.MPAN,
		ProfileClass: data.ProfileClass,
	}

	for _, gsp := range GSPs {
		if gsp.GSPGroupID == data.GspID {
			mPoint.GSP = gsp
			return mPoint, nil
		}
	}

	return MeterPoint{}, errors.New("no grid supply point found")
}

// GetGridSupplyPoint gets a grid supply point based on postcode
// https://developer.octopus.energy/docs/api/#list-grid-supply-points
func (c *Client) GetGridSupplyPoint(postcode string) (GridSupplyPoint, error) {
	// Check if postcode is valid
	if !checkPostcode(postcode) {
		return GridSupplyPoint{}, errors.Errorf("invalid postcode %s", postcode)
	}

	// Remove spaces from postcode
	postcode = strings.ReplaceAll(postcode, " ", "")

	// Struct for JSON unmarshalling
	data := struct {
		Count    int    `json:"count"`
		Next     string `json:"next"`
		Previous string `json:"previous"`
		Results  []struct {
			GroupID string `json:"group_id"`
		} `json:"results"`
	}{}

	err := c.do(fmt.Sprintf("industry/grid-supply-points/?postcode=%s", postcode), &data)
	if err != nil {
		return GridSupplyPoint{}, errors.Errorf("error retrieving grid supply point: %v", err)
	}

	// Only return data if we are dealing with a single result
	if len(data.Results) != 1 {
		return GridSupplyPoint{}, errors.New("more than one supply point received")
	}

	for _, gsp := range GSPs {
		if gsp.GSPGroupID == data.Results[0].GroupID {
			return gsp, nil
		}
	}

	return GridSupplyPoint{}, errors.New("unknown grid supply point")
}

// GetMeterConsumption retrieves meter consumption
// https://developer.octopus.energy/docs/api/#consumption
func (c *Client) GetMeterConsumption(mpan, serialNo string, options ConsumptionOption) ([]Consumption, error) {
	data := struct {
		Count        int           `json:"count"`
		NextPage     string        `json:"next"`
		PreviousPage string        `json:"previous"`
		Results      []Consumption `json:"results"`
	}{}

	apiURL, err := url.Parse(fmt.Sprintf("electricity-meter-points/%s/meters/%s/consumption/", mpan, serialNo))
	if err != nil {
		return nil, errors.Errorf("unable to parse request url: %v", err)
	}

	// Add options to URL if they are provided
	if options != (ConsumptionOption{}) {
		q := apiURL.Query()
		if options.PageSize != 0 {
			q.Add("page_size", strconv.Itoa(options.PageSize))
		}
		if options.OrderBy != "" {
			q.Add("order_by", options.OrderBy)
		}
		if options.GroupBy != "" {
			q.Add("group_by", options.GroupBy)
		}
		if !options.From.IsZero() {
			q.Add("period_from", options.From.Format(iso8601))
		}
		if !options.To.IsZero() {
			q.Add("period_to", options.To.Format(iso8601))
		}
		apiURL.RawQuery = q.Encode()
	}

	err = c.do(apiURL.String(), &data)
	if err != nil {
		return nil, errors.Errorf("error retrieving meter consumption: %v", err)
	}

	return data.Results, nil
}

// checkPostcode checks if provided string is a valid UK postcode
func checkPostcode(postcode string) bool {
	return postcodeRegex.MatchString(postcode)
}

// listProductsPage retrieves products from a single page of JSON data
func (c *Client) listProductsPage(URL string) ([]Product, string, error) {
	var data productJSON

	err := c.do(URL, &data)
	if err != nil {
		return nil, "", errors.Errorf("error retrieving: %v", err)
	}

	return data.Results, data.Next, nil
}

// ListProducts returns a list of energy products
// https://developer.octopus.energy/docs/api/#list-products
func (c *Client) ListProducts() ([]Product, error) {
	var products []Product

	URL := fmt.Sprintf("%s/products/", c.URL)

	for {
		pageProducts, url, err := c.listProductsPage(URL)
		URL = url
		if err != nil {
			return nil, errors.Errorf("error retrieving products page: %v", err)
		}

		for _, product := range pageProducts {
			products = append(products, product)
		}

		if URL == "" {
			break
		}
	}

	return products, nil
}

// GetProduct retrieves a product based on its name
// https://developer.octopus.energy/docs/api/#retrieve-a-product
func (c *Client) GetProduct(productCode string) (Product, error) {
	var product Product

	err := c.do(fmt.Sprintf("products/%s/", productCode), &product)
	if err != nil {
		return Product{}, errors.Errorf("error retrieving the product: %v", err)
	}

	return product, nil
}

// urlAddUsername adds username to URL
func urlAddUsername(URL, username string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", errors.Errorf("error parsing url: %v", err)
	}

	u.User = url.UserPassword(username, "")
	return u.String(), nil
}

func (c *Client) do(path string, v interface{}) error {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/%s", c.URL, path))
	if err != nil {
		return errors.Errorf("http get error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http error - code %d received", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return errors.Errorf("unable to unmarshal json: %v", err)
	}

	return nil
}
