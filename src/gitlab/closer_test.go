package gitlab

import (
	"gitlab-mr-autocloser/src/gitlab/mocks"
	"testing"
)

//func setup(t *testing.T) (*http.ServeMux, *gitlab.Client) {
//	// mux is the HTTP request multiplexer used with the test server.
//	mux := http.NewServeMux()
//
//	// server is a test HTTP server used to provide mock API responses.
//	server := httptest.NewServer(mux)
//	t.Cleanup(server.Close)
//
//	// client is the Gitlab client being tested.
//	//client, err := gitlab.NewClient("",
//	//	gitlab.WithBaseURL(server.URL),
//	//	// Disable backoff to speed up tests that expect errors.
//	//	gitlab.WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
//	//		return 0
//	//	}),
//	//)
//
//	log := logger.SetupLogger(true)
//
//	client, err := gitlab.NewClient("",
//		gitlab.WithBaseURL(server.URL),
//		gitlab.WithCustomLogger(log),
//	)
//
//	if err != nil {
//		t.Fatalf("Failed to create client: %v", err)
//	}
//
//	return mux, client
//}

//func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
//	f, err := os.Open(fixturePath)
//	if err != nil {
//		t.Fatalf("error opening fixture file: %v", err)
//	}
//
//	if _, err = io.Copy(w, f); err != nil {
//		t.Fatalf("error writing response: %v", err)
//	}
//}

func TestGetOpenMRs(t *testing.T) {
	//mux, client := setup(t)
	//
	//path := "/api/v4/projects/test-group1/test-project1/merge_requests"
	//
	//mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
	//	mustWriteHTTPResponse(t, w, "testdata/merge_requests.json")
	//})

	//log := logrus.New()
	//
	//cfg := config.AutoCloserConfig{
	//	DefaultOptions: config.DefaultOptions{
	//		StaleMRAfterDays: 14,
	//		CloseMRAfterDays: 7,
	//	},
	//	Projects: []config.ProjectConfigs{
	//		{
	//			Name:            "test-group1/test-project1",
	//			OverrideOptions: config.OverrideOptions{},
	//		},
	//	},
	//}

	//client := new(gitlab.Client)

	mrc := mocks.NewMRCloser(t)

	mrc.Test(t)

	//mergeRequests := mrc.
	//	require.Equal(t, 3, len(*mergeRequests))
}
