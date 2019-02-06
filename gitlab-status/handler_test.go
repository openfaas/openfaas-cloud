package function

import (
	"fmt"
	"testing"
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
			title:       "URL with path",
			eventURL:    "https://some.random.url/some/routing/here",
			SHA:         "99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			id:          3,
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			expectedErr: nil,
		},
		{
			title:       "URL is empty",
			eventURL:    "",
			SHA:         "99a7c6009c43cca39c61977ab8d3abdd13d7b111",
			id:          3,
			expectedURL: "",
			expectedErr: fmt.Errorf("eventURL or SHA are empty"),
		},
		{
			title:       "SHA is empty",
			eventURL:    "https://some.random.url/some/routing/here",
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
		expectedURL string
	}{
		{
			title:       "Everything is set right",
			url:         "https://some.random.url/api/v4/projects/3/statuses/99a7c6009",
			state:       "pending",
			desc:        "some description",
			context:     "stack-deploy",
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=pending",
		},
		{
			title:       "Everything is set right and state is failure",
			url:         "https://some.random.url/api/v4/projects/3/statuses/99a7c6009",
			state:       "failure",
			desc:        "some description",
			context:     "stack-deploy",
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=failed",
		},
		{
			title:       "Showing that already existing parameters are overwritten",
			url:         "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?value=somevalue",
			state:       "failure",
			desc:        "some description",
			context:     "context",
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?context=context&description=some+description&state=failed",
		},
		{
			title:       "When state is empty",
			url:         "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?value=somevalue",
			state:       "",
			desc:        "some description",
			context:     "stack-deploy",
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?context=stack-deploy&description=some+description&state=",
		},
		{
			title:       "When everything is empty",
			url:         "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?value=somevalue",
			state:       "",
			desc:        "",
			context:     "",
			expectedURL: "https://some.random.url/api/v4/projects/3/statuses/99a7c6009?context=&description=&state=",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			wholeUrl, _ := appendParameters(test.url, test.state, test.desc, test.context)
			if wholeUrl != test.expectedURL {
				t.Errorf("wanted: %s got: %s", test.expectedURL, wholeUrl)
			}
		})
	}
}
