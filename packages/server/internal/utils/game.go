package utils

func CalculateAccuracy(score, totalAttempts int32) float32 {
	accuracy := float32(0)
	if totalAttempts > 0 {
		accuracy = float32(score) / float32(totalAttempts)
	}

	return accuracy * 100
}

func MaskSessionID(s string) string {
	length := len(s)
	return "****" + s[length-4:]
}
