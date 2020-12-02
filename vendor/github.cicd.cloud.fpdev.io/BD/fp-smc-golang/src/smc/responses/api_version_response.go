package responses

type ApiVersionResponse struct {
	Version []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"version"`
}
