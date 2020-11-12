package test

import "testing"

func Test_CoreOFBuilderDep_NonHttpProbe(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=false",
		"--set", "edgeAuth.enableOAuth2=false",
		"--set", "ofBuilder.replicas=5",
		"--set", "global.enableECR=true",
		"--set", "buildKit.privileged=false",
	}

	want := makeOFBuilderDep(false, true, 5, "false", "moby/buildkit:v0.6.2")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/of-builder-dep.yaml", want, t)
}

func makeOFBuilderDep(httpProbe, isECR bool, replicas int, buildkitPrivileged, ofBuildkitImage string) YamlSpec {
	labels := make(map[string]string)

	labels["app.kubernetes.io/component"] = "of-builder"
	labels["app.kubernetes.io/instance"] = "RELEASE-NAME"
	labels["app.kubernetes.io/managed-by"] = "Helm"
	labels["app.kubernetes.io/name"] = "openfaas-cloud"
	labels["helm.sh/chart"] = "openfaas-cloud-0.12.1"

	deployVolumes := makeOFBuilderDeployVolumes([]string{"registry-secret", "payload-secret"}, isECR)
	containerVolumes := makeOFBuilderContainerVolumes(isECR)
	containerEnvironment := makeOFBuilderContainerEnv()

	var readinessProbe LivenessProbe
	if httpProbe {
		readinessProbe = LivenessProbe{
			HttpGet: HttpProbe{
				Path: "/healthz",
				Port: 8080,
			},
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			InitialDelaySeconds: 2,
		}
	} else {
		readinessProbe = LivenessProbe{
			ExecProbe: ExecProbe{
				Command: []string{"wget", "--quiet", "--tries=1", "--timeout=5", "--spider", "http://localhost:8080/healthz"},
			},
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			InitialDelaySeconds: 2,
		}
	}
	return YamlSpec{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: MetadataItems{
			Name:      "of-builder",
			Namespace: "openfaas",
			Labels:    labels,
		},
		Spec: Spec{
			Replicas: replicas,
			Selector: MatchLabelSelector{MatchLabels: map[string]string{"app": "of-builder"}},
			Template: SpecTemplate{
				Metadata: MetadataItems{
					Annotations: map[string]string{"prometheus.io.scrape": "false"},
					Labels:      map[string]string{"app": "of-builder"},
				},
				Spec: TemplateSpec{
					Volumes: deployVolumes,
					Containers: []DeploymentContainers{{
						Name:                    "of-builder",
						Image:                   "openfaas/of-builder:0.8.0",
						ImagePullPolicy:         "IfNotPresent",
						ContainerReadinessProbe: readinessProbe,
						ContainerEnvironment:    containerEnvironment,
						Ports: []ContainerPort{{
							Port:     8080,
							Protocol: "TCP",
						}},
						Volumes: containerVolumes,
					},
						{
							Name:            "of-buildkit",
							Args:            []string{"--addr", "tcp://0.0.0.0:1234"},
							Image:           ofBuildkitImage,
							ImagePullPolicy: "IfNotPresent",
							Ports: []ContainerPort{{
								Port:     1234,
								Protocol: "TCP",
							}},
							SecurityContext: map[string]string{"privileged": buildkitPrivileged},
						}},
				},
			},
		},
	}
}

func makeOFBuilderContainerVolumes(isECR bool) []ContainerVolume {
	var vols []ContainerVolume
	vols = append(vols, ContainerVolume{
		Name:      "registry-secret",
		ReadOnly:  true,
		MountPath: "/home/app/.docker/",
	})

	vols = append(vols, ContainerVolume{
		Name:      "payload-secret",
		ReadOnly:  true,
		MountPath: "/var/openfaas/secrets/",
	})

	if isECR {
		vols = append(vols, ContainerVolume{
			Name:      "aws-ecr-credentials",
			ReadOnly:  true,
			MountPath: "/home/app/.aws/",
		})
	}

	return vols
}

func makeOFBuilderContainerEnv() []Environment {
	var environ []Environment
	environ = append(environ, Environment{Name: "enable_lchown", Value: "true"})
	environ = append(environ, Environment{Name: "insecure", Value: "false"})
	environ = append(environ, Environment{Name: "buildkit_url", Value: "tcp://127.0.0.1:1234"})
	environ = append(environ, Environment{Name: "disable_hmac", Value: "false"})

	return environ
}

func makeOFBuilderDeployVolumes(names []string, isECR bool) []DeploymentVolumes {
	volumes := []DeploymentVolumes{}

	for name := range names {
		volumes = append(volumes, DeploymentVolumes{
			SecretName: names[name],
			SecretInfo: map[string]string{"secretName": names[name]},
		})
	}

	if isECR {
		volumes = append(volumes, DeploymentVolumes{
			SecretName: "aws-ecr-credentials",
			SecretInfo: map[string]string{"secretName": "aws-ecr-credentials", "defaultMode": "420"},
		})
	}

	return volumes
}
