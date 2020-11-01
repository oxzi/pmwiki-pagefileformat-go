package pmwiki

type patchType int

const (
	_ patchType = iota

	// addition of lines from a Patch.
	addition
	// deletion of lines from a Patch.
	deletion
	// change of lines from a Patch, combined addition and deletion.
	change
)

type patchAction struct {
	mode          patchType
	startLine     int
	additionLines []string
	deletionLines []string
}

type Patch []patchAction
