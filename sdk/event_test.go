package sdk

import "testing"

func Test_BuildEventFromPushEvent_ForEventKey(t *testing.T) {
	p := PushEvent{
		Ref: "refs/tags/simple-tag",
		Repository: PushEventRepository{
			Name:  "svc",
			Owner: Owner{},
		},
		Installation: PushEventInstallation{},
	}

	event := BuildEventFromPushEvent(p)

	want := "svc-simple-tag"
	if event.EventKey != want {
		t.Errorf("want %s, got %s", want, event.EventKey)
		t.Fail()
	}
}
