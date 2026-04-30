package app

type cleanupStack []func()

func (stack *cleanupStack) Add(cleanup func()) {
	*stack = append(*stack, cleanup)
}

func (stack cleanupStack) Run() <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		for cleanupIndex := len(stack) - 1; cleanupIndex >= 0; cleanupIndex-- {
			stack[cleanupIndex]()
		}
		close(doneChan)
	}()

	return doneChan
}
