package internal

// const lifecycleChaincodeName = "_lifecycle"

func ApplyOptions[T any, O ~func(*T) error](target *T, options ...O) error {
	for _, option := range options {
		if err := option(target); err != nil {
			return err
		}
	}

	return nil
}
