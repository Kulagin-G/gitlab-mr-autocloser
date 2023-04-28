package gitlab

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"gitlab-mr-autocloser/src/config"
	"gitlab-mr-autocloser/src/gitlab/types"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func setup(path string, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *gitlab.Client) {
	mux := http.NewServeMux()
	mux.HandleFunc(path, handler)

	server := httptest.NewServer(mux)

	client, _ := gitlab.NewClient("",
		gitlab.WithBaseURL(server.URL),
	)

	return server, client
}
func TestNew(t *testing.T) {
	t.Log("[TEST]: Check that New method returns mrCloser interface.")

	cfg := config.AutoCloserConfig{
		DefaultOptions: config.DefaultOptions{
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 1,
		},
		Projects: []config.ProjectConfigs{
			{
				Name: "test-group1/test-project1",
			},
			{
				Name: "test-group1/test-project1",
				OverrideOptions: config.OverrideOptions{
					StaleMRAfterDays: 20,
					CloseMRAfterDays: 25,
				},
			},
		},
	}
	log := logrus.New()
	nmc := New(&cfg, log)

	want := mrCloser{
		cfg: &cfg,
		log: log,
	}

	if !reflect.DeepEqual(&want, nmc) {
		t.Errorf("Labels.UpdateLabel returned %+v, want %+v", nmc, &want)
	}
}

func TestGetOpenMRsNil(t *testing.T) {
	t.Log("[TEST]: Check that GetOpenMR works correctly if no open stale MRs")

	mock, client := setup("/api/v4/projects/test-group1/test-project1/merge_requests", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`)
	})
	defer mock.Close()

	cfg := config.AutoCloserConfig{
		DefaultOptions: config.DefaultOptions{
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 1,
		},
		Projects: []config.ProjectConfigs{
			{
				Name: "test-group1/test-project1",
			},
		},
	}

	nmc := New(&cfg, logrus.New())
	openMRs := nmc.GetOpenMRs(client)

	want := &[]types.MRWithMeta{}

	if !reflect.DeepEqual(*want, *openMRs) {
		t.Errorf("Labels.UpdateLabel returned %+v, want %+v", openMRs, want)
	}
}

func TestGetOpenMRs(t *testing.T) {
	t.Log("[TEST]: Check that staleMRAfterDays and closeMRAfterDays are setting correctly.")

	mock, client := setup("/api/v4/projects/test-group1/test-project1/merge_requests", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"id":1, "iid": 1, "project_id" : 1}]`)
	})
	defer mock.Close()

	cfg := config.AutoCloserConfig{
		DefaultOptions: config.DefaultOptions{
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 1,
		},
		Projects: []config.ProjectConfigs{
			{
				Name: "test-group1/test-project1",
			},
			{
				Name: "test-group1/test-project1",
				OverrideOptions: config.OverrideOptions{
					StaleMRAfterDays: 20,
					CloseMRAfterDays: 25,
				},
			},
		},
	}

	nmc := New(&cfg, logrus.New())
	openMRs := nmc.GetOpenMRs(client)

	want := &[]types.MRWithMeta{
		{
			ProjectID:        1,
			ProjectName:      "test-group1/test-project1",
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 1,
			OpenMR: &gitlab.MergeRequest{
				ID:        1,
				IID:       1,
				ProjectID: 1,
			},
		},
		{
			ProjectID:        1,
			ProjectName:      "test-group1/test-project1",
			StaleMRAfterDays: 20,
			CloseMRAfterDays: 25,
			OpenMR: &gitlab.MergeRequest{
				ID:        1,
				IID:       1,
				ProjectID: 1,
			},
		},
	}

	if !reflect.DeepEqual(*want, *openMRs) {
		t.Errorf("Labels.UpdateLabel returned %+v, want %+v", openMRs, want)
	}
}

func TestSetLabelMR(t *testing.T) {
	t.Log("[TEST]: Check that new MR label is setting correctly.")

	mock, client := setup("/api/v4/projects/1/merge_requests/1000", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":1, "iid": 1, "project_id" : 1, "labels": ["testLabelKey::99"]}`)
	})
	defer mock.Close()

	cfg := config.AutoCloserConfig{
		LabelHead: "testLabelKey::",
	}

	mr := types.MRWithMeta{
		ProjectID:        1,
		CloseMRAfterDays: 99,
		OpenMR: &gitlab.MergeRequest{
			ID:  1,
			IID: 1000,
		},
	}

	nmc := New(&cfg, logrus.New())
	_, label, err := nmc.SetLabelMR(client, &mr)

	if err != nil {
		t.Fatal(err)
	}

	want := "testLabelKey::99"

	if !reflect.DeepEqual(want, label) {
		t.Errorf("Labels.UpdateLabel returned %+v, want %+v", label, want)
	}
}

func TestCloseMRs(t *testing.T) {
	t.Log("[TEST]: Check that CloseMRs method closes stale MRs correctly.")

	mock, client := setup("/api/v4/projects/1/merge_requests/10", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":1, "iid": 10, "project_id": 1}`)
	})
	defer mock.Close()

	cfg := config.AutoCloserConfig{
		LabelHead: "closeLabelKey::",
	}
	mrs := []types.MRWithMeta{
		{
			ProjectID:        1,
			ProjectName:      "DONT_CLOSE_1",
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 7,
			OpenMR: &gitlab.MergeRequest{
				IID:       10,
				WebURL:    "DONT_CLOSE_1",
				UpdatedAt: func() *time.Time { t := time.Now().Add(-5 * 24 * time.Hour); return &t }(),
				Labels: gitlab.Labels{
					"closeLabelKey::7",
				},
			},
		},
		{
			ProjectID:        1,
			ProjectName:      "DONT_CLOSE_2",
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 25,
			OpenMR: &gitlab.MergeRequest{
				IID:       10,
				WebURL:    "DONT_CLOSE_2",
				UpdatedAt: func() *time.Time { t := time.Now().Add(-20 * 24 * time.Hour); return &t }(),
				Labels:    gitlab.Labels{},
			},
		},
		{
			ProjectID:        1,
			ProjectName:      "TO_CLOSE_1",
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 9,
			OpenMR: &gitlab.MergeRequest{
				IID:       10,
				WebURL:    "TO_CLOSE_1",
				UpdatedAt: func() *time.Time { t := time.Now().Add(-10 * 24 * time.Hour); return &t }(),
				Labels: gitlab.Labels{
					"closeLabelKey::9",
				},
			},
		},
		{
			ProjectID:        1,
			ProjectName:      "TO_CLOSE_2",
			StaleMRAfterDays: 5,
			CloseMRAfterDays: 18,
			OpenMR: &gitlab.MergeRequest{
				IID:       10,
				WebURL:    "TO_CLOSE_2",
				UpdatedAt: func() *time.Time { t := time.Now().Add(-20 * 24 * time.Hour); return &t }(),
				Labels: gitlab.Labels{
					"closeLabelKey::19",
				},
			},
		},
	}

	nmc := New(&cfg, logrus.New())
	closedMRs := nmc.CloseMRs(client, &mrs)

	assert.Equal(t, len(closedMRs), 2)
}

func TestManageMergeRequests(t *testing.T) {
	t.Log("[TEST]: Check that MRCloser client is being created correctly.")

	cfg := config.AutoCloserConfig{
		GitlabApiToken:   "12345",
		GitlabBaseApiUrl: "test.com",
	}
	err := New(&cfg, logrus.New()).ManageMergeRequests()

	assert.NoError(t, err)
}
