package function

import (
	"fmt"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Test_gitLabURLBuilder(t *testing.T) {
	tests := []struct {
		title       string
		eventURL    string
		SHA         string
		id          int
		expectedURL string
		expectedErr error
	}{
		{
			title:       "URL is populated and it has path",
			eventURL:    "https://www.random.xyz/some/routing/here",
			SHA:         "99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			id:          3,
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			expectedErr: nil,
		},
		{
			title:       "URL is empty but the SHA is populated",
			eventURL:    "",
			SHA:         "99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			id:          3,
			expectedURL: "",
			expectedErr: fmt.Errorf("eventURL or SHA are empty"),
		},
		{
			title:       "URL is populated but SHA is empty",
			eventURL:    "https://www.random.xyz/some/routing/here",
			SHA:         "",
			id:          3,
			expectedURL: "",
			expectedErr: fmt.Errorf("eventURL or SHA are empty"),
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			url, urlErr := gitLabURLBuilder(test.eventURL, test.SHA, test.id)
			if url != test.expectedURL {
				t.Errorf("expected: %s got: %s", test.expectedURL, url)
			}
			if urlErr == nil {
				if urlErr != test.expectedErr {
					t.Errorf("expected: %v got: %v", test.expectedErr, urlErr)
				}
			} else {
				if urlErr.Error() != test.expectedErr.Error() {
					t.Errorf("expected: %v got: %v", test.expectedErr, urlErr)
				}
			}
		})
	}
}

func Test_appendParameters(t *testing.T) {
	tests := []struct {
		title       string
		url         string
		state       string
		desc        string
		context     string
		targetURL   string
		expectedURL string
	}{
		{
			title:       "Every parameter is set the proper way",
			url:         "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009",
			state:       "pending",
			desc:        "some description",
			context:     "stack-deploy",
			targetURL:   "https://www.world.com/",
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=pending&target_url=https%3A%2F%2Fwww.world.com%2F",
		},
		{
			title:       "Every parameter is set right and state is failure",
			url:         "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009",
			state:       "failure",
			desc:        "some description",
			context:     "stack-deploy",
			targetURL:   "https://www.world.com/",
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=failed&target_url=https%3A%2F%2Fwww.world.com%2F",
		},
		{
			title:       "This test shows that if the URL already has parameters they will be overwritten",
			url:         "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?value=somevalue&target_url=https://www.world.com/",
			state:       "failure",
			desc:        "some description",
			context:     "context",
			targetURL:   "https://www.world.com/",
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?context=context&description=some+description&state=failed&target_url=https%3A%2F%2Fwww.world.com%2F",
		},
		{
			title:       "State parameter is empty, everything other parameter is set right",
			url:         "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?value=somevalue",
			state:       "",
			desc:        "some description",
			context:     "stack-deploy",
			targetURL:   "https://www.world.com/",
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=&target_url=https%3A%2F%2Fwww.world.com%2F",
		},
		{
			title:       "Every parameter is empty, but URL exists",
			url:         "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?value=somevalue",
			state:       "",
			desc:        "",
			context:     "",
			targetURL:   "",
			expectedURL: "https://www.random.xyz/api/v4/projects/3/statuses/99a7c6009?context=&description=&state=&target_url=",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			wholeUrl, _ := appendParameters(test.url, test.state, test.desc, test.context, test.targetURL)
			if wholeUrl != test.expectedURL {
				t.Errorf("wanted: %s got: %s", test.expectedURL, wholeUrl)
			}
		})
	}
}

func Test_setTargetURL(t *testing.T) {
	tests := []struct {
		title              string
		eventService       string
		functionStatus     string
		functionContext    string
		availableRedirects Endpoints
		expectedtargetURL  string
	}{
		{
			title:           "Status is Success so the targetURL points to function endpoint",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusSuccess,
			functionContext: "user-cloud",
			availableRedirects: Endpoints{
				FunctionEndpointURL: "https://user.o6s.io/cloud",
			},
			expectedtargetURL: "https://user.o6s.io/cloud",
		},
		{
			title:           "Status is Failure which means that the targetURL will point to logs",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusFailure,
			functionContext: "user-cloud",
			availableRedirects: Endpoints{
				LogsURL: "https://system.o6s.io/pipeline-logs",
			},
			expectedtargetURL: "https://system.o6s.io/pipeline-logs",
		},
		{
			title:           "Status is Pending which means that the targetURL will point to logs",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusPending,
			functionContext: "user-cloud",
			availableRedirects: Endpoints{
				LogsURL: "https://system.o6s.io/pipeline-logs",
			},
			expectedtargetURL: "https://system.o6s.io/pipeline-logs",
		},
		{
			title:           "The when the context is stack-deploy it should always point to the dashboard, when status is Success",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusSuccess,
			functionContext: sdk.StackContext,
			availableRedirects: Endpoints{
				DashboardURL: "https://system.o6s.io/dashboard/user",
			},
			expectedtargetURL: "https://system.o6s.io/dashboard/user",
		},
		{
			title:           "The when the context is stack-deploy it should always point to the dashboard, when status is Pending",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusPending,
			functionContext: sdk.StackContext,
			availableRedirects: Endpoints{
				DashboardURL: "https://system.o6s.io/dashboard/user",
			},
			expectedtargetURL: "https://system.o6s.io/dashboard/user",
		},
		{
			title:           "The when the context is stack-deploy it should always point to the dashboard, when status is Failure",
			eventService:    "user-cloud",
			functionStatus:  sdk.StatusPending,
			functionContext: sdk.StackContext,
			availableRedirects: Endpoints{
				DashboardURL: "https://system.o6s.io/dashboard/user",
			},
			expectedtargetURL: "https://system.o6s.io/dashboard/user",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			targetURL := setTargetURL(test.functionStatus, test.functionContext, test.eventService, test.availableRedirects)
			if targetURL != test.expectedtargetURL {
				t.Errorf("expected URL: %s got: %s", test.expectedtargetURL, targetURL)
			}
		})
	}
}

func Test_getTargetURLs(t *testing.T) {
	/*
	* Since we already test the functions inside
	* this function we check only if the function
	* populates the variable in the right way
	 */

	event := &sdk.Event{
		Owner:      "user",
		Service:    "cloud-func",
		Repository: "cloud-func-repo",
		SHA:        "98zxc823axcc",
	}
	gateway := "https://system.o6s.io/"
	expectedURLs := Endpoints{
		LogsURL:             "https://system.o6s.io/dashboard/user/cloud-func/log?repoPath=user/cloud-func-repo&commitSHA=98zxc823axcc",
		FunctionEndpointURL: "https://user.o6s.io/cloud-func",
		DashboardURL:        "https://system.o6s.io/dashboard/user",
	}
	targetURLs, urlErr := getTargetURLs(gateway, event)
	if urlErr != nil {
		t.Errorf("expected error to be nil, got: %s", urlErr.Error())
	}
	if expectedURLs.LogsURL != targetURLs.LogsURL {
		t.Errorf("expected the logs URL to be: %s got: %s", expectedURLs.LogsURL, targetURLs.LogsURL)
	}
	if expectedURLs.DashboardURL != targetURLs.DashboardURL {
		t.Errorf("expected the dashboard URL to be: %s got: %s", expectedURLs.DashboardURL, targetURLs.DashboardURL)
	}
	if expectedURLs.FunctionEndpointURL != targetURLs.FunctionEndpointURL {
		t.Errorf("expected the dashboard URL to be: %s got: %s", expectedURLs.FunctionEndpointURL, targetURLs.FunctionEndpointURL)
	}
}
