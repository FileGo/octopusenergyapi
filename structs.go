package octopusenergyapi

import (
	"net/http"
	"time"
)

const (
	iso8601         = "2006-01-02T15:04:05.000+0000"
	baseURL         = "https://api.octopus.energy/v1"
	fuelElectricity = "electricity"
	fuelGas         = "gas"
)

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

// GridSupplyPoint represents a Grid Supply Point (GSP)
type GridSupplyPoint struct {
	ID            int
	Name          string
	Operator      string
	PhoneNumber   string
	ParticipantID string
	GSPGroupID    string
}

// Client represents a Client to be used with the API
type Client struct {
	httpClient *http.Client
	URL        string
}

// MeterPoint represents a meter point
// https://developer.octopus.energy/docs/api/#retrieve-a-meter-point
type MeterPoint struct {
	GSP          GridSupplyPoint
	MPAN         string
	ProfileClass int
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
