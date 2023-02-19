package nixsearch

type Input struct {
	Channel string
	Query   string
}

type Output struct {
	Input       *Input
	Derivations []string
}

func Search(input Input) (*Output, error) {
	dummyData := Output{
		Input:       &input,
		Derivations: []string{"gcloud"},
	}

	return &dummyData, nil
}
