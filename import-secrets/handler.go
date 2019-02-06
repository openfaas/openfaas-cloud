package function

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"

	hmac "github.com/alexellis/hmac"
	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	ssv1alpha1clientset "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealed-secrets/v1alpha1"
	"github.com/openfaas/openfaas-cloud/sdk"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handle uses SealedSecrets to bind the secrets
// to the cluster so the buildshiprun can bind user
// secrets to function
func Handle(req []byte) string {
	event := getEventFromHeader()

	if sdk.HmacEnabled() {
		key, err := sdk.ReadSecret("payload-secret")
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)

			return err.Error()
		}

		digest := os.Getenv("Http_X_Cloud_Signature")

		validated := hmac.Validate(req, digest, key)

		if validated != nil {
			fmt.Fprintf(os.Stderr, validated.Error())
			os.Exit(1)

			return "Unable to validate HMAC"
		}
		fmt.Fprintf(os.Stderr, "hash for HMAC validated successfully for %s\n", event.owner)
	}

	config, err := rest.InClusterConfig()

	if err != nil {
		fmt.Println("couldn't get cluster config - err:", err)
		os.Exit(-1)
	}

	ssc := ssv1alpha1clientset.NewForConfigOrDie(config)

	var userSecret SealedSecret

	err = yaml.Unmarshal(req, &userSecret)

	if err != nil {
		fmt.Println("couldn't unmarshall secrets.yml\n", err)
		os.Exit(-1)
	}

	if len(event.owner) == 0 {
		return fmt.Sprintf("invalid owner name %s", event.owner)
	}

	if strings.HasPrefix(userSecret.Metadata.Name, event.owner) == false {
		return fmt.Errorf("unable to bind a secret which does not start with owner name: %s", event.owner).Error()
	}

	name := fmt.Sprintf("%s", userSecret.Metadata.Name)

	existingSS, err := ssc.SealedSecrets(userSecret.Metadata.Namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		fmt.Printf("found SealedSecret %s start updating\n", name)

		err = updateEncryptedData(existingSS, &userSecret)
		_, err := ssc.SealedSecrets(userSecret.Metadata.Namespace).Update(existingSS)

		if err != nil {
			fmt.Printf("couldn't update SealedSecret (%s) - error: %s\n", name, err)
			os.Exit(-1)
		}
		return fmt.Sprintf("Imported SealedSecret: %s via update", name)
	}

	if !errors2.IsNotFound(err) {
		fmt.Printf("couldn't get SealedSecret (%s) - error: %s\n", name, err)
		os.Exit(-1)
	}

	var ss ssv1alpha1.SealedSecret

	ss.Name = name
	ss.ObjectMeta = *userSecret.Metadata
	ss.ObjectMeta.Name = name
	ss.Spec = ssv1alpha1.SealedSecretSpec{
		EncryptedData: map[string][]byte{},
	}

	err = updateEncryptedData(&ss, &userSecret)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	_, createErr := ssc.SealedSecrets(ss.Namespace).Create(&ss)

	if createErr != nil {
		fmt.Printf("couldn't create SealedSecret (%s) - error: %s\n", name, createErr)
		os.Exit(-1)
	}

	return fmt.Sprintf("Imported SealedSecret: %s as new object", name)
}

func updateEncryptedData(ss *ssv1alpha1.SealedSecret, userSecret *SealedSecret) error {
	for k, v := range userSecret.Spec.EncryptedData {
		encodedBytes, err := base64.StdEncoding.DecodeString(v)

		if err != nil {
			return fmt.Errorf("can't decode base64 string (%s) - error: %s", k, err)
		}

		ss.Spec.EncryptedData[k] = encodedBytes
	}

	return nil
}

func getEventFromHeader() *eventInfo {
	info := eventInfo{
		owner: os.Getenv("Http_Owner"),
	}

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
