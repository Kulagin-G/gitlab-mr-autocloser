package gitlab

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	gl "github.com/xanzy/go-gitlab"
	"gitlab-mr-autocloser/src/config"
	"gitlab-mr-autocloser/src/gitlab/types"
	"net/http"
	"strings"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.0 --name=MRCloser
type MRCloser interface {
	GetOpenMRs(c *gl.Client) *[]types.MRWithMeta
	SetLabelMR(c *gl.Client, mr *types.MRWithMeta)
	CloseMRs(c *gl.Client, mrs *[]types.MRWithMeta)
	ManageMergeRequests()
}

type mrCloser struct {
	cfg *config.AutoCloserConfig
	log *logrus.Logger
}

func NewMRCloser(cfg *config.AutoCloserConfig, log *logrus.Logger) MRCloser {
	cl := mrCloser{
		cfg: cfg,
		log: log,
	}
	return &cl
}

func (mc *mrCloser) GetOpenMRs(c *gl.Client) *[]types.MRWithMeta {
	var mrsWithMeta []types.MRWithMeta
	var opts gl.ListProjectMergeRequestsOptions

	for _, pr := range mc.cfg.Projects {
		// If pr.OverrideOptions.StaleMRAfterDays is not set then default is 0
		if pr.OverrideOptions.StaleMRAfterDays > 0 {
			createdBefore := time.Now().Add(time.Duration(-pr.OverrideOptions.StaleMRAfterDays) * 24 * time.Hour)
			opts = gl.ListProjectMergeRequestsOptions{
				State:         gl.String("opened"),
				CreatedBefore: &createdBefore,
			}
		} else {
			createdBefore := time.Now().Add(time.Duration(-mc.cfg.DefaultOptions.StaleMRAfterDays) * 24 * time.Hour)
			opts = gl.ListProjectMergeRequestsOptions{
				State:         gl.String("opened"),
				CreatedBefore: &createdBefore,
			}
		}

		mc.log.Infof("Checking %s project...", pr.Name)
		mr, _, err := c.MergeRequests.ListProjectMergeRequests(pr.Name, &opts)
		if err != nil {
			mc.log.Errorf("Can't fetch MRs from %s project: %v", pr.Name, err)
		} else {

			var staleMRAfterDays, closeMRAfterDays int

			if pr.OverrideOptions.StaleMRAfterDays == 0 {
				staleMRAfterDays = mc.cfg.DefaultOptions.StaleMRAfterDays
			}
			if pr.OverrideOptions.CloseMRAfterDays == 0 {
				closeMRAfterDays = mc.cfg.DefaultOptions.CloseMRAfterDays
			}
			for _, m := range mr {
				mrsWithMeta = append(mrsWithMeta, types.MRWithMeta{
					ProjectID:        m.ProjectID,
					ProjectName:      pr.Name,
					OpenMR:           m,
					StaleMRAfterDays: staleMRAfterDays,
					CloseMRAfterDays: closeMRAfterDays,
				})
			}
		}

	}
	return &mrsWithMeta
}

func (mc *mrCloser) SetLabelMR(c *gl.Client, mr *types.MRWithMeta) {
	label := fmt.Sprintf("%s%d", mc.cfg.LabelHead, mr.CloseMRAfterDays)
	opts := &gl.UpdateMergeRequestOptions{
		AddLabels: &gl.Labels{
			label,
		},
	}
	_, _, err := c.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.OpenMR.IID, opts)
	if err != nil {
		mc.log.Errorf("Label '%s' was not added to merge request %d: %v", label, mr.OpenMR.IID, err)
	} else {
		mc.log.Infof("Label '%s' added to merge request %d", label, mr.OpenMR.IID)
	}

}

func (mc *mrCloser) CloseMRs(c *gl.Client, mrs *[]types.MRWithMeta) {
	for _, mr := range *mrs {
		toClose := false
		for _, l := range mr.OpenMR.Labels {
			if ok := strings.HasPrefix(l, mc.cfg.LabelHead); ok == true {
				toClose = true
			}
		}
		if toClose {
			sinceLastUpdatesDays := time.Now().Sub(*mr.OpenMR.UpdatedAt).Hours() / 24
			mc.log.Infof("MR %d from %s is already stale, no updates: %f days", mr.OpenMR.IID, mr.ProjectName, sinceLastUpdatesDays)

			if sinceLastUpdatesDays > float64(mr.CloseMRAfterDays) {
				opts := &gl.UpdateMergeRequestOptions{
					StateEvent: gl.String("close"),
				}
				_, _, err := c.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.OpenMR.IID, opts)
				if err != nil {
					mc.log.Errorf("Merge request %s was not closed: %v", mr.OpenMR.WebURL, err)
				} else {
					mc.log.Infof("Merge request %s has been closed!", mr.OpenMR.WebURL)
				}
			}
		} else {
			mc.log.Infof("Found a stale MR %d in %s!", mr.OpenMR.IID, mr.ProjectName)
			mc.SetLabelMR(c, &mr)
		}
	}
}

func (mc *mrCloser) ManageMergeRequests() {
	mc.log.Info("Starting CRON task...")

	client, err := gl.NewClient(mc.cfg.GitlabApiToken,
		gl.WithBaseURL(mc.cfg.GitlabBaseApiUrl),
		gl.WithCustomRetryMax(5),
		gl.WithCustomLogger(mc.log),
		gl.WithCustomRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if resp != nil && (resp.StatusCode == 500 || resp.StatusCode == 503) {
				return true, nil
			}
			return false, nil
		}),
	)

	if err != nil {
		mc.log.Errorf("Error occured during Gitlab client creation: %v\n", err)
	} else {
		openMRs := mc.GetOpenMRs(client)
		mc.CloseMRs(client, openMRs)
	}
}
