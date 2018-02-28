package loop

import (
	"testing"
)

func tt(a, b interface{}) bool {
	return a == b
}
func TestQueue(t *testing.T) {
	q := NewQueue()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.Add("d")

	ok := tt(q.Peek(), "a")

	ok = ok && tt(q.Get(0), "a")
	ok = ok && tt(q.Top(), "a")
	ok = ok && tt(q.Top(), "b")
	ok = ok && tt(q.Top(), "c")
	ok = ok && tt(q.Top(), "d")
	ok = ok && tt(q.Top(), nil)
	ok = ok && tt(q.Peek(), nil)

	if !ok {
		t.Error("loop test error")
	}

}
