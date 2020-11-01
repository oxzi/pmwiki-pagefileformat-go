package pmwiki

type patchType int

const (
	_ patchType = iota

	// Addition of lines from a Patch.
	Addition
	// Deletion of lines from a Patch.
	Deletion
	// Change of lines from a Patch, combined Addition and Deletion.
	Change
)

type patchAction struct {
	mode          patchType
	startLine     int
	additionLines []string
	deletionLines []string
}

type Patch []patchAction
