package api

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

type Connections struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Connections   []Connection `json:"connections"`
}

type Metadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationIP   string `json:"destinationIP"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	InboundName     string `json:"inboundName"`
	InboundUser     string `json:"inboundUser"`
}

type Connection struct {
	ID       string   `json:"id"`
	Network  string   `json:"network"`
	Upload   int64    `json:"upload"`
	Download int64    `json:"download"`
	Metadata Metadata `json:"metadata"`
}
