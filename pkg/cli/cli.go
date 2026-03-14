package cli

type Mode string

const (
	Download Mode = `download`
	Upload   Mode = `upload`
	Delete   Mode = `delete`
	Stream   Mode = `stream`
)

type Flags struct {
	Mode Mode `json:"mode"`

	// Download, Upload, Delete, Stream
	// Just the subdomain eg: Domain: test.auenkr.qzz.io => test
	// Download: used as index file
	// Upload,Delete, Stream: used to create index file
	Subdomain string `json:"subdomain"`

	// Download and Upload
	// Download: Save location of file that will be downloaded
	// Upload: Path of file that will be uploaded
	Path string `json:"path"` // Download path
}
