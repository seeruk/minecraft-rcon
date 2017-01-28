package errhandling

type Stack struct {
	errors []error
}

func NewStack() *Stack {
	return new(Stack)
}

func (s *Stack) Add(error error) {
	s.errors = append(s.errors, error)
}

func (s *Stack) Empty() bool {
	return len(s.errors) == 0
}
