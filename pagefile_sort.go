package pmwiki

// ByTime implements sort.Interface for []PageFile based on PageFile.Time.
type ByTime []PageFile

// Len of the PageFiles.
func (pfs ByTime) Len() int {
	return len(pfs)
}

// Less is true iff one PageFile's Time is before another ones.
func (pfs ByTime) Less(i, j int) bool {
	return pfs[i].Time.Before(pfs[j].Time)
}

// Swap the position of two PageFiles.
func (pfs ByTime) Swap(i, j int) {
	pfs[i], pfs[j] = pfs[j], pfs[i]
}
