package har

import "sort"

func (h *HAR) Merge(another *HAR, compFunction func(e1, e2 *Entry) bool) (*HAR, error) {

	var entries []*Entry
	entries = append(entries, h.Log.Entries...)
	entries = append(entries, another.Log.Entries...)

	sort.SliceStable(entries, func(p, q int) bool {
		return compFunction(entries[p], entries[q])
	})

	merged := HAR{
		Log: &Log{
			Version: h.Log.Version,
			Creator: h.Log.Creator,
			Browser: h.Log.Browser,
			Pages:   nil,
			Entries: entries,
			Comment: h.Log.Comment,
			TraceId: h.Log.TraceId,
		},
	}

	return &merged, nil
}
