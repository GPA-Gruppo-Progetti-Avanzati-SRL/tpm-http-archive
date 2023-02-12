package har

type BuilderOption func(hrab *Builder)

type Builder struct {
	creator        string
	creatorVersion string
	comment        string
	entries        []*Entry
}

func WithEntry(e *Entry) BuilderOption {
	return func(hrab *Builder) {
		hrab.entries = append(hrab.entries, e)
	}
}

func WithCreator(creator string, version string) BuilderOption {
	return func(hrab *Builder) {
		hrab.creator = creator
		hrab.creatorVersion = version
	}
}

func WithComment(comment string) BuilderOption {
	return func(hrab *Builder) {
		hrab.comment = comment
	}
}

func NewHAR(opts ...BuilderOption) *HAR {

	harb := Builder{creator: "rest-client", creatorVersion: "1.0"}
	for _, o := range opts {
		o(&harb)
	}

	har := HAR{
		Log: &Log{
			Version: "1.1",
			Creator: &Creator{
				Name:    harb.creator,
				Version: harb.creatorVersion,
			},
			Comment: harb.comment,
		},
	}

	har.Log.Entries = append(har.Log.Entries, harb.entries...)
	return &har
}
