package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAGSSyncService_Enabled(t *testing.T) {
	svc := NewAGSSyncService(nil)
	assert.NotNil(t, svc)
	assert.False(t, svc.Enabled())
}

func TestCalculateCompositeScore(t *testing.T) {
	cases := []struct {
		name            string
		score           int32
		accuracy        float32
		durationSeconds int32
		expected        int64
	}{
		{
			name:            "zero values",
			score:           0,
			accuracy:        0,
			durationSeconds: 0,
			expected:        0,
		},
		{
			name:            "score dominates accuracy",
			score:           1,
			accuracy:        100,
			durationSeconds: 0,
			expected:        1000000 + 1000000,
		},
		{
			name:            "accuracy breaks tie on equal score",
			score:           1,
			accuracy:        50,
			durationSeconds: 0,
			expected:        1000000 + 500000,
		},
		{
			name:            "faster finish breaks tie on equal score and accuracy",
			score:           1,
			accuracy:        50,
			durationSeconds: 10,
			expected:        1000000 + 500000 - 10,
		},
		{
			name:            "negative duration is clamped",
			score:           0,
			accuracy:        0,
			durationSeconds: -5,
			expected:        0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := calculateCompositeScore(tc.score, tc.accuracy, tc.durationSeconds)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
