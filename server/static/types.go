package static

type StaticFile struct {
	Name     string
	Path     string
	Ext      string
	Content  []byte
	MimeType string
}
