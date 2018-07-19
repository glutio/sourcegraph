package git

import (
	"os"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// ModeSubmodule is an os.FileMode mask indicating that the file is a Git submodule.
//
// To avoid being reported as a regular file mode by (os.FileMode).IsRegular, it sets other bits
// (os.ModeDevice) beyond the Git "160000" commit mode bits. The choice of os.ModeDevice is
// arbitrary.
const ModeSubmodule os.FileMode = 0160000 | os.ModeDevice

// SubmoduleInfo holds information about a Git submodule and is
// returned in the FileInfo's Sys field by Stat/Lstat/ReadDir calls.
type SubmoduleInfo struct {
	// URL is the submodule repository origin URL.
	URL string

	// CommitID is the pinned commit ID of the submodule (in the
	// submodule repository's commit ID space).
	CommitID api.CommitID
}