package services

func setupTestWordService() *WordService {
	ws := &WordService{
		words: []WordDefinition{
			{Word: "hello"},
			{Word: "world"},
			{Word: "apple"},
			{Word: "banana"},
			{Word: "cherry"},
			{Word: "dragon"},
			{Word: "elephant"},
			{Word: "forest"},
			{Word: "guitar"},
			{Word: "horizon"},
		},
	}
	return ws
}
