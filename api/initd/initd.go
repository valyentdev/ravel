package initd

type FSEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
}

type MkdirOptions struct {
	Dir string `json:"dir"`
}

type Status struct {
	Ok bool `json:"ok"`
}

type WatchFSEvent struct {
	Path   string `json:"path"`
	Create bool   `json:"create"`
	Write  bool   `json:"write"`
	Remove bool   `json:"remove"`
	Rename bool   `json:"rename"`
}
