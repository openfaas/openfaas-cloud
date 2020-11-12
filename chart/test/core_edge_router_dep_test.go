package test

import (
	"testing"
)

func Test_CoreEdgeRouterDep_NonHttpProbe(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=false",
		"--set", "edgeAuth.enableOAuth2=false",
	}

	want := makeEdgeRouterDep(false, false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-router-dep.yaml", want, t)
}

func Test_CoreEdgeRouterDep_HttpProbe(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=true",
		"--set", "edgeAuth.enableOAuth2=false",
	}

	want := makeEdgeRouterDep(true, false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-router-dep.yaml", want, t)
}
func Test_CoreEdgeRouterDep_HttpAndOauth(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=true",
		"--set", "edgeAuth.enableOAuth2=true",
	}

	want := makeEdgeRouterDep(true, true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-router-dep.yaml", want, t)
}

func Test_CoreEdgeRouterDep_NoHttpOauth(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=false",
		"--set", "edgeAuth.enableOAuth2=true",
	}

	want := makeEdgeRouterDep(false, true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-router-dep.yaml", want, t)
}

func makeEdgeRouterDep(httpProbe, oauthEnabled bool) YamlSpec {
	labels := make(map[string]string)

	labels["app.kubernetes.io/component"] = "edge-router"
	labels["app.kubernetes.io/instance"] = "RELEASE-NAME"
	labels["app.kubernetes.io/managed-by"] = "Helm"
	labels["app.kubernetes.io/name"] = "openfaas-cloud"
	labels["helm.sh/chart"] = "openfaas-cloud-0.12.1"

	containerEnvironment := makeRouterContainerEnv(oauthEnabled)

	var readinessProbe LivenessProbe
	if httpProbe {
		readinessProbe = LivenessProbe{
			HttpGet: HttpProbe{
				Path: "/healthz",
				Port: 8080,
			},
			TimeoutSeconds:      5,
			InitialDelaySeconds: 2,
			PeriodSeconds:       10,
		}
	} else {
		readinessProbe = LivenessProbe{
			ExecProbe: ExecProbe{
				Command: []string{"wget", "--quiet", "--tries=1", "--timeout=5", "--spider", "http://localhost:8080/healthz"},
			},
			TimeoutSeconds:      5,
			InitialDelaySeconds: 2,
			PeriodSeconds:       10,
		}
	}
	return YamlSpec{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: MetadataItems{
			Name:      "edge-router",
			Namespace: "openfaas",
			Labels:    labels,
		},
		Spec: Spec{
			Replicas: 1,
			Selector: MatchLabelSelector{MatchLabels: map[string]string{"app": "edge-router"}},
			Template: SpecTemplate{
				Metadata: MetadataItems{
					Annotations: map[string]string{"prometheus.io.scrape": "false"},
					Labels:      map[string]string{"app": "edge-router"},
				},
				Spec: TemplateSpec{
					Containers: []DeploymentContainers{{
						Name:                    "edge-router",
						Image:                   "openfaas/edge-router:0.7.4",
						ImagePullPolicy:         "IfNotPresent",
						ContainerReadinessProbe: readinessProbe,
						ContainerEnvironment:    containerEnvironment,
						Ports: []ContainerPort{{
							Port:     8080,
							Protocol: "TCP",
						}},
					}},
				},
			},
		},
	}
}

func makeRouterContainerEnv(oauthEnabled bool) []Environment {
	var environ []Environment
	environ = append(environ, Environment{Name: "upstream_url", Value: "http://gateway.openfaas:8080"})
	environ = append(environ, Environment{Name: "port", Value: "8080"})
	environ = append(environ, Environment{Name: "timeout", Value: "60s"})

	if oauthEnabled {
		environ = append(environ, Environment{Name: "auth_url", Value: "http://edge-auth.openfaas:8080"})
	} else {
		environ = append(environ, Environment{Name: "auth_url", Value: "http://echo.openfaas-fn:8080"})
	}

	return environ
}
