package traffic

type Snapshot struct {
	UploadBytes   int64 `json:"upload_bytes"`
	DownloadBytes int64 `json:"download_bytes"`
	UploadRate    int64 `json:"upload_rate"`
	DownloadRate  int64 `json:"download_rate"`
	Connections   int   `json:"connections"`
}

type UserTraffic struct {
	UserID       uint  `json:"user_id"`
	UploadBytes  int64 `json:"upload_bytes"`
	DownloadBytes int64 `json:"download_bytes"`
	Online       bool  `json:"online"`
}
