// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package pod

import (
	"github.com/elastic/cloud-on-k8s/operators/pkg/apis/kibana/v1alpha1"
	"github.com/elastic/cloud-on-k8s/operators/pkg/utils/stringsutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// HTTPPort is the (default) port used by Kibana
	HTTPPort                             = 5601
	elasticsearchUsername                = "ELASTICSEARCH_USERNAME"
	elasticsearchPassword                = "ELASTICSEARCH_PASSWORD"
	defaultImageRepositoryAndName string = "docker.elastic.co/kibana/kibana"
)

// DefaultResources are resource limits to apply to Kibana container by default
var DefaultResources = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
}

// ApplyToEnv applies any auth information in auth to the variables in env.
func ApplyToEnv(auth v1alpha1.ElasticsearchAuth, env []corev1.EnvVar) []corev1.EnvVar {
	if auth.Inline != nil {
		env = append(
			env,
			corev1.EnvVar{Name: elasticsearchUsername, Value: auth.Inline.Username},
			corev1.EnvVar{Name: elasticsearchPassword, Value: auth.Inline.Password},
		)
	}
	if auth.SecretKeyRef != nil {
		env = append(
			env,
			corev1.EnvVar{Name: elasticsearchUsername, Value: auth.SecretKeyRef.Key},
			corev1.EnvVar{Name: elasticsearchPassword, ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: auth.SecretKeyRef,
			}})
	}
	return env
}

type SpecParams struct {
	Version          string
	ElasticsearchUrl string
	CustomImageName  string
	User             v1alpha1.ElasticsearchAuth
	PodTemplate      corev1.PodTemplateSpec
}

func imageWithVersion(image string, version string) string {
	return stringsutil.Concat(image, ":", version)
}

type EnvFactory func(p SpecParams) []corev1.EnvVar

func NewSpec(p SpecParams, env EnvFactory) corev1.PodSpec {
	imageName := p.CustomImageName
	if p.CustomImageName == "" {
		imageName = imageWithVersion(defaultImageRepositoryAndName, p.Version)
	}

	probe := &corev1.Probe{
		FailureThreshold:    3,
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      5,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromInt(HTTPPort),
				Path:   "/",
				Scheme: corev1.URISchemeHTTP,
			},
		},
	}

	automountServiceAccountToken := false

	return corev1.PodSpec{
		Affinity: p.PodTemplate.Spec.Affinity,
		Containers: []corev1.Container{{
			Resources: resourceRequirements(p.PodTemplate),
			Env:       env(p),
			Image:     imageName,
			Name:      v1alpha1.KibanaContainerName,
			Ports: []corev1.ContainerPort{
				{Name: "http", ContainerPort: int32(HTTPPort), Protocol: corev1.ProtocolTCP},
			},
			ReadinessProbe: probe,
		}},
		AutomountServiceAccountToken: &automountServiceAccountToken,
	}
}

// resourceRequirements parses the given podTemplate to return Kibana container resource requirements.
// If not set in the podTemplate, returns the default ones.
func resourceRequirements(podTemplate corev1.PodTemplateSpec) corev1.ResourceRequirements {
	for _, c := range podTemplate.Spec.Containers {
		if c.Name == v1alpha1.KibanaContainerName && (len(c.Resources.Limits) > 0 || len(c.Resources.Requests) > 0) {
			return c.Resources
		}
	}
	return DefaultResources
}
