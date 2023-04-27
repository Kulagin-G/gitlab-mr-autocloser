package types

import (
	"github.com/xanzy/go-gitlab"
)

type MRWithMeta struct {
	ProjectID        int
	ProjectName      string
	StaleMRAfterDays int
	CloseMRAfterDays int
	OpenMR           *gitlab.MergeRequest
}
