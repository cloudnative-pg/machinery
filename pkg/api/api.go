package api

import corev1 "k8s.io/api/core/v1"

// SecretKeySelector contains enough information to let you locate
// the key of a Secret
type SecretKeySelector struct {
	// The name of the secret in the pod's namespace to select from.
	LocalObjectReference `json:",inline"`
	// The key to select
	Key string `json:"key"`
}

// LocalObjectReference contains enough information to let you locate a
// local object with a known type inside the same namespace
type LocalObjectReference struct {
	// Name of the referent.
	Name string `json:"name"`
}

// ConfigMapKeySelector contains enough information to let you locate
// the key of a ConfigMap
type ConfigMapKeySelector struct {
	// The name of the secret in the pod's namespace to select from.
	LocalObjectReference `json:",inline"`
	// The key to select
	Key string `json:"key"`
}

// SecretKeySelectorToCore transforms a SecretKeySelector structure to the
// analogue one in the corev1 namespace
func SecretKeySelectorToCore(selector *SecretKeySelector) *corev1.SecretKeySelector {
	if selector == nil {
		return nil
	}

	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: selector.LocalObjectReference.Name,
		},
		Key: selector.Key,
	}
}

// ConfigMapKeySelectorToCore transforms a ConfigMapKeySelector structure to the analogue
// one in the corev1 namespace
func ConfigMapKeySelectorToCore(selector *ConfigMapKeySelector) *corev1.ConfigMapKeySelector {
	if selector == nil {
		return nil
	}

	return &corev1.ConfigMapKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: selector.Name,
		},
		Key: selector.Key,
	}
}