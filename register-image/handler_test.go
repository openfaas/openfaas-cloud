package function

import "testing"

func Test_GetImage_StripsRegistryAndTag(t *testing.T) {

	r := ImageRequest{Image: "docker.io/alexellis/fn:latest"}

	got := r.GetRepository()
	want := "alexellis/fn"

	if got != want {
		t.Errorf("GetRepository want: %s, got %s", want, got)
		t.Fail()
	}
}

func Test_GetImage_OFCExample(t *testing.T) {

	r := ImageRequest{Image: "docker.io/ofcommunity/alexellis-double-espresso-mug:0.3-master-439d4aa"}

	got := r.GetRepository()
	want := "ofcommunity/alexellis-double-espresso-mug"

	if got != want {
		t.Errorf("GetRepository want: %s, got %s", want, got)
		t.Fail()
	}
}

func Test_GetImageStripsECRRegistry(t *testing.T) {
	r := ImageRequest{Image: "1238475.dkr.ecr.eu-central-1.amazonaws.com/ofbuilder/go-for-it"}

	got := r.GetRepository()
	want := "ofbuilder/go-for-it"

	if got != want {
		t.Errorf("GetRepository want: %s, got %s", want, got)
		t.Fail()
	}
}
