package types

import (
	gl "github.com/xanzy/go-gitlab"
)

type MRWithMeta struct {
	ProjectID        int
	ProjectName      string
	StaleMRAfterDays int
	CloseMRAfterDays int
	OpenMR           *gl.MergeRequest
}
