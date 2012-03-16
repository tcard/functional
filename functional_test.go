package functional

import (
	"reflect"
	"testing"
)

func TestEquals(t *testing.T) {
	l1 := Link(1, Link(2, Link(3, Link(4, Link(5, Empty)))))
	l2 := Link(1, Link(2, Link(3, Link(4, Link(5, Empty)))))
	if !l1.Equals(l2) {
		t.Errorf("%v.Equals(%v)", l1, l2)
	}
	l2 = Link(1, Link(2, Link(5, Empty)))
	if l1.Equals(l2) {
		t.Errorf("%v.Equals(%v)", l1, l2)
	}
}

func TestList(t *testing.T) {
	l := Link(1, Link(2, Link(3, Link(4, Link("a", Empty)))))
	if !l.Equals(List(1, 2, 3, 4, "a")) || l.Equals(List(1, "a", 9)) {
		t.Errorf("List(%v)", l)
	}
}

func TestSliceToList(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	l := SliceToList(slice)
	if !l.Equals(List(1, 2, 3, 4, 5)) || l.Equals(List(1, "a", 9)) {
		t.Errorf("SliceToList(%v) -> %v", slice, l)
	}
}

func TestToSlice(t *testing.T) {
	l := List(1, 2, "a", 4, 5)
	if !reflect.DeepEqual(l.ToSlice(), []I{1, 2, "a", 4, 5}) ||
		reflect.DeepEqual(l.ToSlice(), []I{1, "b", 0}) {
		t.Errorf("ToSlice(%v)", l)
	}
}

func TestAppend(t *testing.T) {
	l1 := List(1, 2, 3)
	l2 := List(4, 5, 6)
	l3 := List(1, 2, 3, 4, 5, 6)
	if !l1.Append(l2).Equals(l3) {
		t.Errorf("Append(%v, %v) -> %v", l1, l2, l3)
	}
}

func TestIter(t *testing.T) {
	l := List(1, 2, "a", 4, 5)
	s := []I{1, 2, "a", 4, 5}
	i := 0
	for v := range l.Iter() {
		if v != s[i] {
			t.Errorf("Iter(%v, %v) -> %v != s[%v]", l, s, v, i)
		}
		i++
	}
}

func TestLength(t *testing.T) {
	l1 := List(1, 2, 3)
	l2 := List()
	if l1.Length() != 3 {
		t.Errorf("Length(%v) != 3", l1)
	}
	if l2.Length() != 0 {
		t.Errorf("Length(%v) != 0", l2)
	}
}

func TestAt(t *testing.T) {
	l1 := List(1, 2, 3)
	s := l1.ToSlice()
	for k, v := range s {
		if w := l1.At(uint(k)); v != w {
			t.Errorf("At(%v, %v) != s[%v]", l1, k, k)
		}
	}
}

func TestTake(t *testing.T) {
	l1 := List(1, 2, 3)
	if !l1.Take(0).Equals(List()) || !l1.Take(2).Equals(List(1, 2)) ||
		!l1.Take(4).Equals(l1) {
		t.Errorf("Take(%v)", l1)
	}
}

func TestDrop(t *testing.T) {
	l1 := List(1, 2, 3)
	if !l1.Drop(0).Equals(l1) || !l1.Drop(1).Equals(List(2, 3)) ||
		!l1.Drop(4).Equals(List()) {
		t.Errorf("Drop(%v)", l1)
	}
}

var prog *Thunk
var fact *Thunk

func TestProg(t *testing.T) {
	var integersFromN func(int) *Thunk
	integersFromN = func(n int) *Thunk {
		return DelayedLink(n, func() *Thunk { return integersFromN(n + 1) })
	}
	prog = integersFromN(1)
	if p := prog.Take(10); !p.Equals(List(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)) {
		t.Errorf("prog.Take(10) = (%v)", p)
	}
}

func TestFact(t *testing.T) {
	var factN func(int) int
	factN = func(n int) int {
		if n == 0 {
			return 1
		}
		return n * factN(n-1)
	}
	var makeFact func(int) *Thunk
	makeFact = func(n int) *Thunk {
		return DelayedLink(factN(n), func() *Thunk { return makeFact(n + 1) })
	}
	fact = makeFact(0)
	if p := fact.Take(5); !p.Equals(List(1, 1, 2, 6, 24)) {
		t.Errorf("fact.Take(10) = (%v) %v", p, L(1, 1, 2, 6, 24))
	}
}

func TestMapN(t *testing.T) {
	m := MapN(func(xs ...I) I {
		switch xs[0].(type) {
		case int:
			r := 0
			for _, v := range xs {
				r += v.(int)
			}
			return r
		}
		return xs
	}, List(1, 2, 2), List(3, 9, 3, 5), prog)
	if l := List(5, 13, 8); !m.Equals(l) {
		t.Errorf("MapN = %v", m)
	}
}

func TestMap(t *testing.T) {
	m := prog.Map(func(x I) I {
		return x.(int) * 2
	})
	if l := List(2, 4, 6, 8, 10); !m.Take(5).Equals(l) {
		t.Errorf("MapN = %v", m)
	}
}

func TestReduceN(t *testing.T) {
	totalSum := func(acc I, xs ...I) I {
		for _, x := range xs {
			acc = acc.(int) + x.(int)
		}
		return acc
	}
	r := ReduceN(totalSum, 0, L(7, 8, 9), L(4, 5, 6), prog)
	if i := 45; r != i {
		t.Errorf("ReduceN = %v != %v", r, i)
	}
}

func TestReduce(t *testing.T) {
	sum := func(acc I, x I) I {
		return acc.(int) + x.(int)
	}
	r := prog.Take(10).Reduce(sum, 0)
	if i := 10 * (10 + 1) / 2; r != i {
		t.Errorf("Reduce = %v != %v", r, i)
	}
}

func TestFilterN(t *testing.T) {
	growing := func(xs ...I) bool {
		prev := xs[0].(int)
		for _, i := range xs[1:] {
			if i.(int) < prev {
				return false
			}
		}
		return true
	}
	r := FilterN(growing, L(1, 2, 3), L(2, 1, 3), L(3, 0, 3))
	if l := L(L(1, 2, 3), L(3, 3, 3)); !r.Equals(l) {
		t.Errorf("FilterN = %v", r)
	}
}

var evens *Thunk

func TestFilter(t *testing.T) {
	evens = prog.Filter(func(x I) bool {
		return x.(int)%2 == 0
	})
	if l := L(2, 4, 6, 8, 10); !l.Equals(evens.Take(5)) {
		t.Errorf("Filter = %v", evens.Take(5))
	}
}

func TestAny(t *testing.T) {
	if evens.Take(20).Any(func(x I) bool {
		return x.(int)%2 == 1
	}) {
		t.Error()
	}
	if !L(1, 7, 26, 54, 20).Any(func(x I) bool {
		return x.(int)%13 == 0
	}) {
		t.Error()
	}
}

func TestAll(t *testing.T) {
	f := func(x I) bool {
		return x.(int)%2 == 0
	}
	if e := evens.Take(20); !e.All(f) || L(2, 4, 10, 3, 6).All(f) {
		t.Error()
	}
}

func TestHas(t *testing.T) {
	if l := L(1, 2, 3, 4, 5); !l.Has(3) || l.Has(6) {
		t.Error()
	}
}

func TestMax(t *testing.T) {
	if l := L(1, 3, 5, 4, 2); l.Max().(int) != 5 {
		t.Errorf("%v", l.Max())
	}
	if l := L(1, 3, 5.0, 4, 2); l.Max() != nil {
		t.Errorf("%v", l.Max())
	}
	if l := L([]int{1, 2, 3}, 3, 5, 4, 2); l.Max() != nil {
		t.Errorf("%v", l.Max())
	}
}

func TestMin(t *testing.T) {
	if l := L(3, 2, 5, 1, 2); l.Min().(int) != 1 {
		t.Errorf("%v", l.Min())
	}
	if l := L(1, 3, 5.0, 4, 2); l.Min() != nil {
		t.Errorf("%v", l.Min())
	}
	if l := L([]int{1, 2, 3}, 3, 5, 4, 2); l.Min() != nil {
		t.Errorf("%v", l.Min())
	}
}

func TestTakeWhile(t *testing.T) {
	lt6 := func(x I) bool {
		return x.(int) < 6
	}
	if l := L(1, 2, 3, 4, 5); !l.Equals(prog.TakeWhile(lt6)) {
		t.Error()
	}
}

func TestDropWhile(t *testing.T) {
	lt3 := func(x I) bool {
		return x.(int) < 3
	}
	if l := L(3, 4, 5); !l.Equals(prog.DropWhile(lt3).Take(3)) {
		t.Errorf("%v", prog.DropWhile(lt3).Take(3))
	}
}

func TestZipN(t *testing.T) {
	if l := L(L(1, 2), L(2, 4), L(3, 6)); !l.Equals(ZipN(prog, evens).Take(3)) {
		t.Error()
	}
}

func TestZip(t *testing.T) {
	if l := L(L(1, 2), L(2, 4), L(3, 6)); !l.Equals(prog.Zip(evens).Take(3)) {
		t.Error()
	}
}

func TestFlatten(t *testing.T) {
	if l := L(1, 2, 3, 4); !l.Equals(L(prog.Take(2), L(3, 4)).Flatten()) {
		t.Error()
	}
}

func TestReverse(t *testing.T) {
	if l := L(5, 4, 3, 2, 1); !l.Equals(prog.Take(5).Reverse()) {
		t.Error()
	}
}

func TestLast(t *testing.T) {
	if L(1, 2, 3, 4, 5).Last() != 5 {
		t.Error()
	}
}

func TestUpdating(t *testing.T) {
	prog = Updating(1, func(x I) I {
		return x.(int) + 1
	})
	if l := L(1, 2, 3, 4, 5); !l.Equals(prog.Take(5)) {
		t.Errorf("%v", prog.Take(5))
	}
}

var fibo *Thunk

func TestFibo(t *testing.T) {
	fibo = Link(1, DelayedLink(1, func() *Thunk {
		return MapN(func(xs ...I) I {
			ret := 0
			for _, v := range xs {
				ret += v.(int)
			}
			return ret
		}, fibo, fibo.Tail())
	}))
	if l := L(1, 1, 2, 3, 5); !l.Equals(fibo.Take(5)) {
		t.Error()
	}
}

func usualFibo(n int) int {
	if n <= 1 {
		return 1
	}
	return usualFibo(n-1) + usualFibo(n-2)
}

func BenchmarkIterSlice(b *testing.B) {
	b.StopTimer()
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range s {
		}
	}
}

func BenchmarkIterList(b *testing.B) {
	b.StopTimer()
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range l.Iter() {
		}
	}
}

func BenchmarkIterListNoMemo(b *testing.B) {
	b.StopTimer()
	StopMemo()
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range l.Iter() {
		}
	}
	StartMemo()
}

func BenchmarkMapLoop(b *testing.B) {
	b.StopTimer()
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	double := func(n int) int {
		return n * 2
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range s {
			s[k] = double(v)
		}
		_ = s[0]
	}
}

func BenchmarkMapList(b *testing.B) {
	b.StopTimer()
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	double := func(n I) I {
		return n.(int) * 2
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Map(double).At(0)
	}
}

func BenchmarkMapListNoMemo(b *testing.B) {
	b.StopTimer()
	StopMemo()
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	double := func(n I) I {
		return n.(int) * 2
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Map(double).At(0)
	}
	StartMemo()
}

func BenchmarkAppendSlices(b *testing.B) {
	b.StopTimer()
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = append(s, s...)
	}
}

func BenchmarkAppendLists(b *testing.B) {
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Append(l)
	}
}

func BenchmarkAppendLsNoMemo(b *testing.B) {
	b.StopTimer()
	StopMemo()
	l := L(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Append(l)
	}
	StartMemo()
}

func BenchmarkUsualFibo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		usualFibo(30)
	}
}

func BenchmarkFiboStream(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fibo.At(30)
	}
}

func BenchmarkFiboStrNoMemo(b *testing.B) {
	b.StopTimer()
	StopMemo()
	fibo = Link(1, DelayedLink(1, func() *Thunk {
		return MapN(func(xs ...I) I {
			ret := 0
			for _, v := range xs {
				ret += v.(int)
			}
			return ret
		}, fibo, fibo.Tail())
	}))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = fibo.At(30)
	}
	StartMemo()
}
