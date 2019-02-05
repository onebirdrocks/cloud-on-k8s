package volume

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestSecretVolumeItemProjection(t *testing.T) {

	testVolume := NewSelectiveSecretVolumeWithMountPath("secret", "secrets", "/mnt", []string{"foo"})
	tests := []struct {
		volume   SecretVolume
		expected []corev1.KeyToPath
	}{
		{
			volume:   NewSecretVolume("secret", "/secrets"),
			expected: nil,
		},
		{
			volume: testVolume,
			expected: []corev1.KeyToPath{
				corev1.KeyToPath{
					Key:  "foo",
					Path: "foo",
				},
			},
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.volume.Volume().Secret.Items)
	}
}