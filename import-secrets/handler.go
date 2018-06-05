package function

import (
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	ssv1alpha1clientset "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealed-secrets/v1alpha1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handle a serverless request
func Handle(req []byte) []byte {
	event := getEvent()
	config, err := rest.InClusterConfig()

	if err != nil {
		fmt.Println("couldn't get cluster config - err:", err)
		os.Exit(-1)
	}

	ssc := ssv1alpha1clientset.NewForConfigOrDie(config)

	var ss SealedSecret

	err = yaml.Unmarshal(req, &ss)

	if err != nil {
		fmt.Println("couldn't unmarshall secrets.yml\n", err)
		os.Exit(-1)
	}

	name := fmt.Sprintf("%s-%s", event.owner, ss.Metadata.Name)

	existingSS, err := ssc.SealedSecrets(ss.Metadata.Namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		if !errors2.IsNotFound(err) {
			fmt.Printf("couldn't get SealedSecret (%s) - error: %s\n", name, err)
			os.Exit(-1)
		}

		var sss ssv1alpha1.SealedSecret

		sss.Name = name
		sss.ObjectMeta = *ss.Metadata
		sss.ObjectMeta.Name = name
		sss.Spec = ssv1alpha1.SealedSecretSpec{
			EncryptedData: map[string][]byte{},
		}

		err = updateEncryptedData(&sss, &ss)

		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		_, err := ssc.SealedSecrets(sss.Namespace).Create(&sss)

		if err != nil {
			fmt.Printf("couldn't create SealedSecret (%s) - error: %s\n", name, err)
			os.Exit(-1)
		}
	} else {
		fmt.Printf("found SealedSecret %s start updating\n", name)

		err = updateEncryptedData(existingSS, &ss)

		_, err := ssc.SealedSecrets(ss.Metadata.Namespace).Update(existingSS)

		if err != nil {
			fmt.Printf("couldn't update SealedSecret (%s) - error: %s\n", name, err)
			os.Exit(-1)
		}
	}

	return []byte(fmt.Sprintf("handled sealed secrets from secrets.yml"))
}

func updateEncryptedData(sss *ssv1alpha1.SealedSecret, css *SealedSecret) error {
	for k, v := range css.Spec.EncryptedData {
		bs, err := base64.StdEncoding.DecodeString(v)

		if err != nil {
			return fmt.Errorf("couldn't decode base64 string (%s) - error: %s\n", k, err)
		}

		sss.Spec.EncryptedData[k] = bs
	}

	return nil
}

func getEvent() *eventInfo {
	info := eventInfo{}

	info.owner = os.Getenv("Http_Owner")

	return &info
}

type eventInfo struct {
	owner string
}

type SealedSecretSpec struct {
	EncryptedData map[string]string `yaml:"encryptedData"`
}

type SealedSecret struct {
	ApiVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Metadata   *metav1.ObjectMeta `yaml:"metadata"`
	Spec       SealedSecretSpec   `yaml:"spec"`
}
