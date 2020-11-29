package octopusenergyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	iso8601 = "2006-01-02T15:04:05.000+0000"
	baseURL = "https://api.octopus.energy/v1"
)

// GridSupplyPoint represents a Grid Supply Point (GSP)
type GridSupplyPoint struct {
	ID            int
	Name          string
	Operator      string
	PhoneNumber   string
	ParticipantID string
	GSPGroupID    string
}

// GSPs provides a list of Grid Supply Points (GSP)
// https://en.wikipedia.org/wiki/Meter_Point_Administration_Number#Distributor_ID
var GSPs = [...]GridSupplyPoint{
	{10, "Eastern England", "UK Power Networks", "0800 029 4285", "EELC", "_A"},
	{11, "East Midlands", "Western Power Distribution", "0800 096 3080", "EMEB", "_B"},
	{12, "London", "UK Power Networks", "0800 029 4285", "LOND", "_C"},
	{13, "Merseyside and Northern Wales", "SP Energy Networks", "0330 10 10 444", "MANW", "_D"},
	{14, "West Midlands", "Western Power Distribution", "0800 096 3080", "MIDE", "_E"},
	{15, "North Eastern England", "Northern Powergrid", "0800 011 3332", "NEEB", "_F"},
	{16, "North Western England", "Electricity North West", "0800 048 1820", "NORW", "_G"},
	{17, "Northern Scotland", "Scottish & Southern Electricity Networks", "0800 048 3516", "HYDE", "_P"},
	{18, "Southern Scotland", "SP Energy Networks", "0330 10 10 444", "SPOW", "_N"},
	{19, "South Eastern England", "UK Power Networks", "0800 029 4285", "SEEB", "_J"},
	{20, "Southern England", "Scottish & Southern Electricity Networks", "0800 048 3516", "SOUT", "_H"},
	{21, "Southern Wales", "Western Power Distribution", "0800 096 3080", "SWAE", "_K"},
	{22, "South Western England", "Western Power Distribution", "0800 096 3080", "SWEB", "_L"},
	{23, "Yorkshire", "Northern Powergrid", "0800 011 3332", "YELG", "_M"},
}

// PCs represents Profile Class of a meter point
// https://en.wikipedia.org/wiki/Meter_Point_Administration_Number#Profile_Class_(PC)
var PCs = map[int]string{
	0: "Half-hourly supply (import and export)",
	1: "Domestic unrestricted",
	2: "Domestic Economy meter of two or more rates",
	3: "Non-domestic unrestricted",
	4: "Non-domestic Economy 7",
	5: "Non-domestic, with maximum demand (MD) recording capability and with load factor (LF) less than or equal to 20%",
	6: "Non-domestic, with MD recording capability and with LF less than or equal to 30% and greater than 20%",
	7: "Non-domestic, with MD recording capability and with LF less than or equal to 40% and greater than 30%",
	8: "Non-domestic, with MD recording capability and with LF greater than 40% (also all non-half-hourly export MSIDs)",
}

// Client represents a Client to be used with the API
type Client struct {
	httpClient *http.Client
	URL        string
}

// NewClient returns a client
func NewClient(APIkey string, httpClient *http.Client) (*Client, error) {
	// Empty APIkey is not permitted
	APIkey = strings.TrimSpace(APIkey)
	if len(APIkey) == 0 {
		return &Client{}, errors.New("API key should not be empty")
	}

	// Add APIkey as username to base URL
	baseURL, err := urlAddUsername(baseURL, APIkey)
	if err != nil {
		return &Client{}, fmt.Errorf("unable to add username to url: %w", err)
	}

	return &Client{
		URL:        baseURL,
		httpClient: httpClient,
	}, nil
}

// MeterPoint represents a meter point
// https://developer.octopus.energy/docs/api/#retrieve-a-meter-point
type MeterPoint struct {
	GSP          GridSupplyPoint
	MPAN         string
	ProfileClass int
}

// GetMeterPoint retrieves a meter point for a given MPAN
// https://developer.octopus.energy/docs/api/#electricity-meter-points
func (c *Client) GetMeterPoint(mpan string) (MeterPoint, error) {
	data := struct {
		GspID        string `json:"gsp"`
		MPAN         string `json:"mpan"`
		ProfileClass int    `json:"profile_class"`
	}{}

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/electricity-meter-points/%s/", c.URL, mpan))
	if err != nil {
		return MeterPoint{}, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MeterPoint{}, fmt.Errorf("http error - code %d received", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return MeterPoint{}, fmt.Errorf("unable to unmarshal json: %w", err)
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
func (c Client) GetGridSupplyPoint(postcode string) (GridSupplyPoint, error) {
	// Check if postcode is valid
	if !checkPostcode(postcode) {
		return GridSupplyPoint{}, fmt.Errorf("invalid postcode %s", postcode)
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

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/industry/grid-supply-points/?postcode=%s", c.URL, postcode))
	if err != nil {
		return GridSupplyPoint{}, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GridSupplyPoint{}, fmt.Errorf("http error - code %d received", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return GridSupplyPoint{}, fmt.Errorf("unable to unmarshal json: %w", err)
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

// Consumption represents a power consumption in a given interval
type Consumption struct {
	// Value represents meter reading for the interval
	// Unit depends on the type of meter:
	//
	// Electricity meters: kWh
	//
	// SMETS1 Secure gas meters: kWh
	//
	// SMETS2 gas meters: m^3
	Value         float32   `json:"consumption"`
	IntervalStart time.Time `json:"interval_start"`
	IntervalEnd   time.Time `json:"interval_end"`
}

// ConsumptionOption represents optional parameters for API.GetMeterConsumption
type ConsumptionOption struct {
	From     time.Time
	To       time.Time
	PageSize int
	OrderBy  string
	GroupBy  string
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

	apiURL, err := url.Parse(fmt.Sprintf("%s/electricity-meter-points/%s/meters/%s/consumption/", c.URL, mpan, serialNo))
	if err != nil {
		return []Consumption{}, fmt.Errorf("unable to parse request url: %w", err)
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

	resp, err := c.httpClient.Get(apiURL.String())
	if err != nil {
		return []Consumption{}, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Consumption{}, fmt.Errorf("http error - code %d received", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return []Consumption{}, fmt.Errorf("unable to unmarshal json: %w", err)
	}

	return data.Results, nil

}

// checkPostcode checks if provided string is a valid UK postcode
func checkPostcode(postcode string) bool {
	match, _ := regexp.MatchString(`^([Gg][Ii][Rr] 0[Aa]{2})|((([A-Za-z][0-9]{1,2})|(([A-Za-z][A-Ha-hJ-Yj-y][0-9]{1,2})|(([AZa-z][0-9][A-Za-z])|([A-Za-z][A-Ha-hJ-Yj-y][0-9]?[A-Za-z])))) [0-9][A-Za-z]{2})$`, postcode)
	return match
}

// Product represents an Octopus Energy product
// https://developer.octopus.energy/docs/api/#retrieve-a-product
type Product struct {
	Code                      string                       `json:"code"`
	Direction                 string                       `json:"direciton"`
	FullName                  string                       `json:"full_name"`
	DisplayName               string                       `json:"display_name"`
	Description               string                       `json:"description"`
	IsVariable                bool                         `json:"is_variable"`
	IsGreen                   bool                         `json:"is_green"`
	IsTracker                 bool                         `json:"is_tracker"`
	IsPrepay                  bool                         `json:"is_prepay"`
	IsBusiness                bool                         `json:"is_business"`
	IsRestricted              bool                         `json:"is_restricted"`
	Term                      int                          `json:"term"`
	AvailableFrom             time.Time                    `json:"available_from"`
	AvailableTo               time.Time                    `json:"available_to"`
	Links                     []Link                       `json:"links"`
	SingleRegisterElecTariffs map[string]map[string]Tariff `json:"single_register_electricity_tariffs"`
	DualRegisterElecTariffs   map[string]map[string]Tariff `json:"dual_register_electricity_tariffs"`
	SingleRegisterGasTariffs  map[string]map[string]Tariff `json:"single_register_gas_tariffs"`
}

// Link represents a hyperlink
type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
	Rel    string `json:"rel"`
}

// Tariff represent an Octopus Energy tariff
type Tariff struct {
	Code                   string  `json:"code"`
	StandingChargeExcVAT   float32 `json:"standing_charge_exc_vat"`
	StandingChargeIncVAT   float32 `json:"standing_charge_inc_vat"`
	OnlineDiscountExcVAT   float32 `json:"online_discount_exc_vat"`
	OnlineDiscountIncVAT   float32 `json:"online_discount_inc_vat"`
	DualFuelDiscountExcVAT float32 `json:"dual_fuel_discount_exc_vat"`
	DualFuelDiscountIncVAT float32 `json:"dual_fuel_discount_inc_vat"`
	ExitFeesExcVAT         float32 `json:"exit_fees_exc_vat"`
	ExitFeesIncVAT         float32 `json:"exit_fees_inc_vat"`
	Links                  []Link  `json:"links"`
	StandardUnitRateExcVAT float32 `json:"standard_unit_rate_exc_vat"`
	StandardUnitRateIncVAT float32 `json:"standard_unit_rate_inc_vat"`
}

type productJSON struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Product `json:"results"`
}

// listProductsPage retrieves products from a single page of JSON data
func (c *Client) listProductsPage(URL string) ([]Product, string, error) {
	var data productJSON

	resp, err := c.httpClient.Get(URL)
	if err != nil {
		return []Product{}, "", fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Product{}, "", fmt.Errorf("http error - code %d received", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return []Product{}, "", fmt.Errorf("unable to unmarshal json: %w", err)
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
			return []Product{}, fmt.Errorf("error retrieving products page: %w", err)
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

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/products/%s/", c.URL, productCode))
	if err != nil {
		return Product{}, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Product{}, fmt.Errorf("http error - code %d received", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		return Product{}, fmt.Errorf("unable to unmarshal json: %w", err)
	}

	return product, nil
}

// urlAddUsername adds username to URL
func urlAddUsername(URL, username string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", fmt.Errorf("error parsing url: %w", err)
	}

	u.User = url.UserPassword(username, "")
	return u.String(), nil
}
