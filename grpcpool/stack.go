package grpcpool

type stack struct {
	values []*stub
}

func (s *stack) Push(stub *stub) {
	s.values = append(s.values, stub)
	return
}

func (s *stack) Pop() (stub *stub) {
	stub = s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return
}

func (s *stack) Peek() *stub {
	return s.values[len(s.values)-1]
}

func (s *stack) Remove(id string) (stub *stub) {
	for i, v := range s.values {
		if v.id == id {
			stub = v
			s.values = append(s.values[:i], s.values[i+1:]...)
			break
		}
	}
	return
}

func (s *stack) Empty() bool {
	return len(s.values) == 0
}

func (s *stack) Size() int {
	return len(s.values)
}
