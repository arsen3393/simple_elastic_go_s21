package types

type Place struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type Limits struct {
	Size int `json:"size"`
	From int `json:"from"`
}
