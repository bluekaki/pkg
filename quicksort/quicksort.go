package quicksort

import (
	"container/list"

	"github.com/bluekaki/pkg/stringutil"
)

type Value interface {
	Compare(Value) stringutil.Diff
}

func Asc(values []Value) {
	if len(values) < 2 {
		return
	}

	stack := list.New()
	stack.PushBack(values)

	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		values := element.Value.([]Value)

		k := 0
		i, j := 0, len(values)-1
		for i <= j {
			for ; j >= i; j-- { // tail to head, find the first less one.
				if diff := values[j].Compare(values[k]); diff == stringutil.Less || diff == stringutil.Equal {
					values[k], values[j] = values[j], values[k]
					k = j
					break
				}
			}

			for ; i <= j; i++ { // head to tail, find the first greater one.
				if diff := values[i].Compare(values[k]); diff == stringutil.Greater {
					values[k], values[i] = values[i], values[k]
					k = i
					break
				}
			}
		}

		if tmp := values[:k]; len(tmp) > 1 {
			stack.PushBack(tmp)
		}

		if tmp := values[k+1:]; len(tmp) > 1 {
			stack.PushBack(tmp)
		}
	}
}

func Desc(values []Value) {
	if len(values) < 2 {
		return
	}

	stack := list.New()
	stack.PushBack(values)

	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		values := element.Value.([]Value)

		k := 0
		i, j := 0, len(values)-1
		for i <= j {
			for ; j >= i; j-- { // tail to head, find the first greater one.
				if diff := values[j].Compare(values[k]); diff == stringutil.Greater || diff == stringutil.Equal {
					values[k], values[j] = values[j], values[k]
					k = j
					break
				}
			}

			for ; i <= j; i++ { // head to tail, find the first less one.
				if diff := values[i].Compare(values[k]); diff == stringutil.Less {
					values[k], values[i] = values[i], values[k]
					k = i
					break
				}
			}
		}

		if tmp := values[:k]; len(tmp) > 1 {
			stack.PushBack(tmp)
		}

		if tmp := values[k+1:]; len(tmp) > 1 {
			stack.PushBack(tmp)
		}
	}
}
