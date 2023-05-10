package gitlab

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"gitlab-mr-autocloser/src/config"
	"gitlab-mr-autocloser/src/gitlab/types"
	"net/http"
	"strings"
	"time"
)

type MRCloser interface {
	GetOpenMRs(c *gitlab.Client) *[]types.MRWithMeta
	SetLabelMR(c *gitlab.Client, mr *types.MRWithMeta) (*gitlab.MergeRequest, string, error)
	CloseMRs(c *gitlab.Client, mrs *[]types.MRWithMeta) []*gitlab.MergeRequest
	ManageMergeRequests() error
}

type mrCloser struct {
	cfg *config.AutoCloserConfig
	log *logrus.Logger
}

func New(cfg *config.AutoCloserConfig, log *logrus.Logger) MRCloser {
	cl := mrCloser{
		cfg: cfg,
		log: log,
	}

	return &cl
}

func (mc *mrCloser) GetOpenMRs(c *gitlab.Client) *[]types.MRWithMeta {
	mrsWithMeta := []types.MRWithMeta{}

	var opts gitlab.ListProjectMergeRequestsOptions

	for _, pr := range mc.cfg.Projects {
		// If pr.OverrideOptions.StaleMRAfterDays is not set then default is 0
		if pr.OverrideOptions.StaleMRAfterDays > 0 {
			createdBefore := time.Now().Add(time.Duration(-pr.OverrideOptions.StaleMRAfterDays) * 24 * time.Hour)
			opts = gitlab.ListProjectMergeRequestsOptions{
				State:         gitlab.String("opened"),
				CreatedBefore: &createdBefore,
				ListOptions: gitlab.ListOptions{
					PerPage: 20,
					Page:    1,
				},
			}
		} else {
			createdBefore := time.Now().Add(time.Duration(-mc.cfg.DefaultOptions.StaleMRAfterDays) * 24 * time.Hour)
			opts = gitlab.ListProjectMergeRequestsOptions{
				State:         gitlab.String("opened"),
				CreatedBefore: &createdBefore,
				ListOptions: gitlab.ListOptions{
					PerPage: 20,
					Page:    1,
				},
			}
		}

		mc.log.Infof("Checking %s project...", pr.Name)

		for {
			mr, resp, err := c.MergeRequests.ListProjectMergeRequests(pr.Name, &opts)

			if err != nil {
				mc.log.Errorf("Can't fetch MRs from %s project: %v", pr.Name, err)
				break
			}

			var staleMRAfterDays, closeMRAfterDays int

			if pr.OverrideOptions.StaleMRAfterDays == 0 {
				staleMRAfterDays = mc.cfg.DefaultOptions.StaleMRAfterDays
			} else {
				staleMRAfterDays = pr.OverrideOptions.StaleMRAfterDays
			}

			if pr.OverrideOptions.CloseMRAfterDays == 0 {
				closeMRAfterDays = mc.cfg.DefaultOptions.CloseMRAfterDays
			} else {
				closeMRAfterDays = pr.OverrideOptions.CloseMRAfterDays
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

			if resp.NextPage == 0 {
				break
			}

			mc.log.Infof("Going to the next page: %v", resp.NextPage)
			opts.Page = resp.NextPage
		}
	}

	return &mrsWithMeta
}

func (mc *mrCloser) SetLabelMR(c *gitlab.Client, mr *types.MRWithMeta) (*gitlab.MergeRequest, string, error) {
	label := fmt.Sprintf("%s%d", mc.cfg.LabelHead, mr.CloseMRAfterDays)
	opts := &gitlab.UpdateMergeRequestOptions{
		AddLabels: &gitlab.Labels{
			label,
		},
	}
	m, _, err := c.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.OpenMR.IID, opts)

	if err != nil {
		mc.log.Errorf("Label '%s' was not added to merge request %d: %v", label, mr.OpenMR.IID, err)

		return m, label, err
	}

	mc.log.Infof("Label '%s' added to merge request %d", label, mr.OpenMR.IID)

	return m, label, nil
}

func (mc *mrCloser) CloseMRs(c *gitlab.Client, mrs *[]types.MRWithMeta) []*gitlab.MergeRequest {
	var closedMRs []*gitlab.MergeRequest

	for _, mr := range *mrs {
		toClose := false

		for _, l := range mr.OpenMR.Labels {
			if ok := strings.HasPrefix(l, mc.cfg.LabelHead); ok {
				toClose = true
			}
		}

		if toClose {
			sinceLastUpdatesDays := time.Since(*mr.OpenMR.UpdatedAt).Hours() / 24
			mc.log.Infof("MR %d from %s is already stale, no updates: %.2f days, threshold: %.2f days", mr.OpenMR.IID, mr.ProjectName, sinceLastUpdatesDays, float64(mr.CloseMRAfterDays))

			if sinceLastUpdatesDays > float64(mr.CloseMRAfterDays) {
				opts := &gitlab.UpdateMergeRequestOptions{
					StateEvent: gitlab.String("close"),
				}

				_, _, err := c.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.OpenMR.IID, opts)

				if err != nil {
					mc.log.Errorf("Merge request %s was not closed: %v", mr.OpenMR.WebURL, err)
				} else {
					mc.log.Infof("Merge request %s has been closed!", mr.OpenMR.WebURL)
					closedMRs = append(closedMRs, mr.OpenMR)
				}
			} else {
				mc.log.Infof("MR %d is not ready to close, will be ready in %.2f days.", mr.OpenMR.IID, float64(mr.CloseMRAfterDays)-sinceLastUpdatesDays)
			}
		} else {
			mc.log.Infof("Found a stale MR %d in %s!", mr.OpenMR.IID, mr.ProjectName)
			mrToLabel := mr
			_, _, _ = mc.SetLabelMR(c, &mrToLabel)
		}
	}

	return closedMRs
}

func (mc *mrCloser) ManageMergeRequests() error {
	mc.log.Info("Starting CRON task...")

	client, err := gitlab.NewClient(mc.cfg.GitlabApiToken,
		gitlab.WithBaseURL(mc.cfg.GitlabBaseApiUrl),
		gitlab.WithCustomRetryMax(5),
		gitlab.WithCustomLogger(mc.log),
		gitlab.WithCustomRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if resp != nil && (resp.StatusCode == 500 || resp.StatusCode == 503) {
				return true, nil
			}

			return false, nil
		}),
	)

	if err != nil {
		mc.log.Errorf("Error occurred during Gitlab client creation: %v\n", err)

		return err
	}

	openMRs := mc.GetOpenMRs(client)
	_ = mc.CloseMRs(client, openMRs)

	return nil
}
