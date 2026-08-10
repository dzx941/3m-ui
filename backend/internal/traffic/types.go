package traffic

type Snapshot struct {
	UploadBytes   int64 `json:"upload_bytes"`
	DownloadBytes int64 `json:"download_bytes"`
	UploadRate    int64 `json:"upload_rate"`
	DownloadRate  int64 `json:"download_rate"`
	Connections   int   `json:"connections"`
}

type UserTraffic struct {
	UserID        uint   `json:"user_id"`
	Username      string `json:"username"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
	TrafficUsed   int64  `json:"traffic_used"`
	TrafficLimit  int64  `json:"traffic_limit"`
	Online        bool   `json:"online"`
}
