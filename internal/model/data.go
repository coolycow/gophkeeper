package model

type Data struct {
	Title      string `json:"title"`
	Login      string `json:"login,omitempty"`
	Password   string `json:"password,omitempty"`
	URL        string `json:"url,omitempty"`
	Notes      string `json:"notes,omitempty"`
	BinaryData []byte `json:"binary_data,omitempty"`
}
