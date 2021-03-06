package atb

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// DefaultURL is the default AtB API URL.
const DefaultURL = "http://st.atb.no/InfoTransit/userservices.asmx"

// Client represents a client which communicates with AtBs API.
type Client struct {
	Username string
	Password string
	URL      string
}

// BusStops represents a list of bus stops.
type BusStops struct {
	Stops []BusStop `json:"Fermate"`
}

// BusStop represents a bus stop.
type BusStop struct {
	StopID      int    `json:"cinFermata"`
	NodeID      string `json:"codAzNodo"`
	Description string `json:"descrizione"`
	Longitude   string `json:"lon"`
	Latitude    int    `json:"lat"`
	MobileCode  string `json:"codeMobile"`
	MobileName  string `json:"nomeMobile"`
}

// Forecasts represents a list of forecasts.
type Forecasts struct {
	Nodes     []NodeInfo `json:"InfoNodo"`
	Forecasts []Forecast `json:"Orari"`
	Total     int        `json:"total"`
}

// NodeInfo represents a bus stop, returned as a part of a forecast.
type NodeInfo struct {
	Name              string `json:"nome_Az"`
	NodeID            string `json:"codAzNodo"`
	NodeName          string `json:"nomeNodo"`
	NodeDescription   string `json:"descrNodo"`
	BitMaskProperties string `json:"bitMaskProprieta"`
	MobileCode        string `json:"codeMobile"`
	Longitude         string `json:"coordLon"`
	Latitude          string `json:"coordLat"`
}

// Forecast represents a single forecast.
type Forecast struct {
	LineID                  string `json:"codAzLinea"`
	LineDescription         string `json:"descrizioneLinea"`
	RegisteredDepartureTime string `json:"orario"`
	ScheduledDepartureTime  string `json:"orarioSched"`
	StationForecast         string `json:"statoPrevisione"`
	Destination             string `json:"capDest"`
}

// NewFromConfig creates a new client where name is the path to the config file.
func NewFromConfig(name string) (Client, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return Client{}, err
	}
	var client Client
	if err := json.Unmarshal(data, &client); err != nil {
		return Client{}, err
	}
	if client.URL == "" {
		client.URL = DefaultURL
	}
	return client, nil
}

func (c *Client) post(m method, data interface{}) ([]byte, error) {
	req, err := m.NewRequest(data)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString(req)
	resp, err := http.Post(c.URL, "application/soap+xml", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsonBlob, err := m.ParseResponse(body)
	if err != nil {
		return nil, err
	}
	return jsonBlob, nil
}

// GetBusStops retrieves bus stops from AtBs API.
func (c *Client) GetBusStops() (BusStops, error) {
	values := struct {
		Username string
		Password string
	}{c.Username, c.Password}

	jsonBlob, err := c.post(busStopsList, values)
	if err != nil {
		return BusStops{}, err
	}

	var stops BusStops
	if err := json.Unmarshal(jsonBlob, &stops); err != nil {
		return BusStops{}, err
	}
	return stops, nil
}

// GetRealTimeForecast retrieves a forecast from AtBs API, using nodeID to
// identify the bus stop.
func (c *Client) GetRealTimeForecast(nodeID int) (Forecasts, error) {
	values := struct {
		Username string
		Password string
		NodeID   int
	}{c.Username, c.Password, nodeID}

	jsonBlob, err := c.post(realTimeForecast, values)
	if err != nil {
		return Forecasts{}, err
	}

	var forecasts Forecasts
	if err := json.Unmarshal(jsonBlob, &forecasts); err != nil {
		return Forecasts{}, err
	}
	return forecasts, nil
}
