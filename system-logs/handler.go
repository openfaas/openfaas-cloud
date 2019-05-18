package function

import (
	"fmt"
	"net/url"
	"os"

	"github.com/openfaas/openfaas-cloud/sdk"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Handle(req []byte) string {

	key, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)

		return err.Error()
	}

	payloadSecret := os.Getenv("Http_X_Cloud_Payload_Secret")

	if key != payloadSecret {
		fmt.Fprintf(os.Stderr, "unauthorized: key X-Cloud-Payload-Secret does not match payload-secret")
		os.Exit(1)
	}

	query := os.Getenv("Http_Query")
	qs, _ := url.ParseQuery(query)
	functionName := qs.Get("function")
	config, err := rest.InClusterConfig()

	if err != nil {
		fmt.Printf("couldn't get cluster config: %s\n", err)
		os.Exit(-1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("couldn't get clientset: %s\n", err)
		os.Exit(-1)
	}

	ns := "openfaas-fn"
	if val, ok := os.LookupEnv("namespace"); ok && len(val) > 0 {
		ns = val
	}

	listOpts := metav1.ListOptions{
		LabelSelector: "faas_function=" + functionName,
	}

	list, err := clientset.CoreV1().Pods(ns).List(listOpts)
	if err != nil {
		fmt.Printf("couldn't list pods in %s: %s\n", ns, err.Error())
		os.Exit(-1)
	}

	if len(list.Items) > 0 {
		pod := list.Items[0]

		ownerLabel := pod.Labels["com.openfaas.cloud.git-owner"]
		if len(ownerLabel) == 0 {
			return fmt.Sprintf("couldn't get logs for non-user function: %s\n", functionName)
		}

		lines := int64(100)
		podLogOpts := &corev1.PodLogOptions{
			Follow:    false,
			TailLines: &lines,
		}

		req := clientset.CoreV1().Pods(ns).GetLogs(pod.Name, podLogOpts)
		res := req.Do()
		rawBytes, _ := res.Raw()

		return string(rawBytes)

	}

	return fmt.Sprintf("No logs available for: %s", functionName)
}
