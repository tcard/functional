// Package functional implements a functional programming library including
// a lazy list implementation and some of the most usual functions.
package functional

import (
	"fmt"
	"reflect"
)

// Type I is the type of the element of a Pair. It is defined as interface{},
// so you can throw anything inside of a Pair. When you take elements back to
// non-functional code you will probably need to type-assert it.
type I interface{}

// The Pair type is the basic element of Lists. Composed of an element (the head)
// and a pointer to the Thunk which returns the next Pair of the List (the tail).
type Pair struct {
	head I
	tail *Thunk
}

type thunkFunc func(I) *Pair

// A Thunk is a delayed Pair. It is a function that, when called, returns 
// or generates, the underlying pair. It takes an argument which may
// be used as a previous value for generating the next Pair, but usually
// it won't take effect. In practice, Thunk is like a Pair which is like a
// List; you won't usually need to worry about the differences.
type Thunk struct {
	id uint64
	f  thunkFunc
}

var thunkId uint64

func nextId() uint64 {
	thunkId++
	return thunkId
}

// Empty is the empty Thunk, that is, a Thunk that returns nil. Lists end
// with it.
var Empty *Thunk = func() *Thunk {
	var v thunkFunc = func(I) *Pair { return nil }
	return &Thunk{nextId(), v}
}()

var memo bool

// Starts memoizing thunk evaluations. By default memoization is on.
func StartMemo() {
	memo = true
}

// Stops memoizing thunk evaluations. By default memoization is on.
func StopMemo() {
	memo = false
}

// Resets the current memoization table. May be useful when it gets too populated
// with values you won't use anymore.
func ResetMemo() {
	memoTable = make(map[uint64]*Pair)
}

var memoTable map[uint64]*Pair

func force(thunk *Thunk, ctx I) *Pair {
	if thunk == nil {
		return nil
	}
	if memo {
		if v, ok := memoTable[(*thunk).id]; ok {
			//fmt.Printf("Yay! %v %v\n", (*thunk).id, v)
			return v
		} else {
			ret := (*thunk).f(ctx)
			//fmt.Printf("Aw... %v %v\n", (*thunk).id, ret)
			memoTable[(*thunk).id] = ret
			return ret
		}
	}
	return (*thunk).f(ctx)
}

func (thunk *Thunk) Head() I {
	return force(thunk, nil).head
}

func (thunk *Thunk) Tail() *Thunk {
	return force(thunk, nil).tail
}

// Takes an element and a Thunk and makes a Thunk with them. Similar
// to Lisp's `cons` or Haskell's `(:)`.
// 	list123 := Link(1, Link(2, Link(3, Empty)))
func Link(head I, tail *Thunk) *Thunk {
	var ret thunkFunc = func(I) *Pair { return &Pair{head, tail} }
	return &Thunk{nextId(), ret}
}

// Performs just like Link, but the tail is doubly delayed. Rarely used,
// useful when the tail is generated by some recursive function.
//
//	var factN func(int) int
//	factN = func(n int) int {
//		if n == 0 {
//			return 1
//		}
//		return n * factN(n-1)
//	}
//	var makeFact func(int) *Thunk
//	makeFact = func(n int) *Thunk {
//		// A direct Link to makeFact(n + 1) would lead to an infinite loop.
//		return DelayedLink(factN(n), func() *Thunk { return makeFact(n + 1) })
//	}
//	fact := makeFact(0)
func DelayedLink(head I, tail func() *Thunk) *Thunk {
	var ret thunkFunc = func(I) *Pair { return &Pair{head, tail()} }
	return &Thunk{nextId(), ret}
}

// Helper function that Links all its arguments. You can easily make a list
// from a slice with it: List(slice...)
func List(items ...I) *Thunk {
	if len(items) >= 1 {
		return Link(items[0], List(items[1:]...))
	}
	return Empty
}

// Shortcut for List.
func L(items ...I) *Thunk {
	return List(items...)
}

func SliceToList(items I) (ret *Thunk) {
	t := reflect.TypeOf(items)
	if items == nil || t.Kind() != reflect.Slice {
		return
	}
	v := reflect.ValueOf(items)
	ret = Empty
	for i := v.Len() - 1; i >= 0; i-- {
		ret = Link(v.Index(i).Interface(), ret)
	}
	return
}

func (thunk *Thunk) ToSlice() [](I) {
	pair := force(thunk, nil)
	if pair == nil {
		return [](I){}
	}
	return append([](I){pair.head}, pair.tail.ToSlice()...)
}

// Makes a single List by appending one to another.
func (thunk *Thunk) Append(other *Thunk) *Thunk {
	var ret thunkFunc = func(_ I) *Pair {
		pair := force(thunk, nil)
		if pair != nil {
			return &Pair{pair.head, pair.tail.Append(other)}
		} else if pair := force(other, nil); pair != nil {
			return &Pair{pair.head, pair.tail.Append(Empty)}
		}
		return nil
	}
	return &Thunk{nextId(), ret}
}

// A handy way of iterating through a List is by calling Iter()
// in a for-range loop.
func (thunk *Thunk) Iter() chan I {
	ch := make(chan I)
	go func() {
		for {
			pair := force(thunk, nil)
			if pair == nil {
				break
			}
			ch <- pair.head
			thunk = pair.tail
		}
		close(ch)
	}()

	return ch
}

func (thunk *Thunk) String() (ret string) {
	ret = "["
	first := true
	for {
		pair := force(thunk, nil)
		if pair == nil {
			ret += "]"
			break
		}
		if !first {
			ret += " "
		} else {
			first = false
		}
		ret += fmt.Sprintf("%v", pair.head)
		thunk = pair.tail
	}
	return
}

// Tests for equality between two lists.
func (thunk *Thunk) Equals(other *Thunk) bool {
	for {
		pair := force(thunk, nil)
		otherPair := force(other, nil)
		if pair == nil {
			if otherPair != nil {
				return false
			} else {
				break
			}
		}
		switch pair.head.(type) {
		case *Thunk:
			switch otherPair.head.(type) {
			case *Thunk:
				if !pair.head.(*Thunk).Equals(otherPair.head.(*Thunk)) {
					return false
				}
			default:
				return false
			}
		default:
			if pair.head != otherPair.head {
				return false
			}
		}
		thunk, other = pair.tail, otherPair.tail
	}
	return true
}

func (thunk *Thunk) Length() (ret int) {
	pair := force(thunk, nil)
	for pair != nil {
		thunk = pair.tail
		pair = force(thunk, nil)
		ret++
	}
	return
}

// Retrieves the element at the n-th position on the list. If there is
// no such element, err is set.
func (thunk *Thunk) At(n uint) (ret I) {
	var pair *Pair
	for i := uint(0); i <= n; i++ {
		pair = force(thunk, nil)
		if pair == nil {
			panic("Index out of list.")
			return
		}
		thunk = pair.tail
	}
	ret = pair.head
	return
}

// Takes the first n elements of a list. Mostly needed for infinite lists.
func (thunk *Thunk) Take(n uint) *Thunk {
	var ret thunkFunc = func(_ I) *Pair {
		if n > 0 {
			pair := force(thunk, nil)
			if pair != nil {
				return &Pair{pair.head, pair.tail.Take(n - 1)}
			}
		}
		return nil
	}
	return &Thunk{nextId(), ret}
}

// Drops the first n elements of a list and returns the rest.
func (thunk *Thunk) Drop(n uint) *Thunk {
	var ret thunkFunc
	ret = func(_ I) *Pair {
		pair := force(thunk, nil)
		if pair != nil {
			if n > 0 {
				n -= 1
				thunk = pair.tail
				return ret(nil)
			}
			return pair
		}
		return nil
	}
	return &Thunk{nextId(), ret}
}

// Applies a function to each element of some lists. The function must
// handle any number of elements. It ends when any of the lists ends.
func MapN(f func(...I) I, thunks ...*Thunk) *Thunk {
	var ret thunkFunc = func(I) *Pair {
		l := len(thunks)
		heads := make([](I), l)
		tails := make([]*Thunk, l)
		for k := 0; k < l; k++ {
			pair := force(thunks[k], nil)
			if pair == nil {
				return nil
			}
			heads[k] = pair.head
			tails[k] = pair.tail
		}
		return &Pair{f(heads...), MapN(f, tails...)}
	}
	return &Thunk{nextId(), ret}
}

// Applies a function to each element of a list.
func (thunk *Thunk) Map(f func(I) I) *Thunk {
	return MapN(func(xs ...I) I {
		return f(xs[0])
	}, thunk)
}

// Applies a function to each element of some lists, returning the
// accumulated value. The function must take the so far accumulated
//  value as its first argument and handle any number of elements as
// the second, third and so on. You pass that initial value as ReduceN's second
// argument. It stops reducing when any of the lists ends.
//
//	func sumInts(xs I...) I {
//		return ReduceN(func(x I, xs ...I) {
//			ret := 0
//			for _, x := range xs {
//				ret += x.(int)
//			}
//			return ret
//		}, 0, xs...)
//	}
//	
// 	sumInts(L(1, 2), L(3, 4)) // == 10 
func ReduceN(f func(I, ...I) I, acc I, thunks ...*Thunk) I {
	l := len(thunks)
	heads := make([](I), l)
	tails := make([]*Thunk, l)
	for k := 0; k < l; k++ {
		pair := force(thunks[k], nil)
		if pair == nil {
			return acc
		}
		heads[k] = pair.head
		tails[k] = pair.tail
	}
	return ReduceN(f, f(acc, heads...), tails...)
}

// Applies a function to each element of a list, returning the
// accumulated value. The function must take the so far accumulated value
// as its first argument and the next element of the list as its second
// one. You pass that initial value as Reduce's second argument.
// It stops reducing when any of the lists ends.
//
//	func (xs *Thunk) sumInts() I {
//		return xs.Reduce(func(acc, x I) {
//				acc = acc.(int) + x.(int)
//			}
//			return acc
//		}, 0)
//	}
//	
// 	L(1,2,3,4).sumInts() // == 10 
func (thunk *Thunk) Reduce(f func(I, I) I, initial I) I {
	return ReduceN(func(acc I, xs ...I) I {
		return f(acc, xs[0])
	}, initial, thunk)
}

// Returns the list of lists of the elements which pass a testing function.
// The testing function must take an element from each list to which
// it is applied.
func FilterN(f func(...I) bool, thunks ...*Thunk) *Thunk {
	var ret thunkFunc = func(I) *Pair {
		l := len(thunks)
		heads := make([](I), l)
		tails := make([]*Thunk, l)
		for k := 0; k < l; k++ {
			pair := force(thunks[k], nil)
			if pair == nil {
				return nil
			}
			heads[k] = pair.head
			tails[k] = pair.tail
		}
		tail := FilterN(f, tails...)
		if f(heads...) {
			return &Pair{L(heads...), tail}
		}
		return force(tail, nil)
	}
	return &Thunk{nextId(), ret}
}

// Returns the lists of the elements of the list that pass a testing
// function.
func (thunk *Thunk) Filter(f func(I) bool) *Thunk {
	/*
		// Slow, but left here for fanciness.
		return FilterN(func(xs ...I) bool {
			return f(xs[0])
		}, thunk).Map(func(arg I) I {
			return arg.(*Thunk).ToSlice()[0]
		})
	*/

	var ret thunkFunc = func(I) *Pair {
		pair := force(thunk, nil)
		if pair == nil {
			return nil
		}
		tail := pair.tail.Filter(f)
		if f(pair.head) {
			return &Pair{pair.head, tail}
		}
		//return (*tail)(tail)
		return force(tail, nil)
	}
	return &Thunk{nextId(), ret}
}

// Tests if any of the elements of the list passes a testing
// function.
func (thunk *Thunk) Any(f func(I) bool) bool {
	pair := force(thunk, nil)
	if pair != nil {
		if f(pair.head) {
			return true
		} else {
			return pair.tail.Any(f)
		}
	}
	return false
}

// Tests if all of the elements of the list passes a testing
// function.
func (thunk *Thunk) All(f func(I) bool) bool {
	pair := force(thunk, nil)
	if pair != nil {
		if f(pair.head) {
			return pair.tail.All(f)
		} else {
			return false
		}
	}
	return true
}

// Tests if a list contains an element.
func (thunk *Thunk) Has(x I) bool {
	return thunk.Any(func(y I) bool {
		return reflect.DeepEqual(x, y)
	})
}

// Retrieves the maximum element of a list. Obviously, the list must be
// composed of ordered elements (ints, floats or strings).
func (thunk *Thunk) Max() I {
	return thunk.Reduce(func(acc, x I) I {
		accV := reflect.ValueOf(acc)
		xV := reflect.ValueOf(x)
		if accV.Kind() != xV.Kind() {
			return nil
		}
		switch accV.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if accV.Int() >= xV.Int() {
				return acc
			} else {
				return x
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			if accV.Uint() >= xV.Uint() {
				return acc
			} else {
				return x
			}
		case reflect.Float32, reflect.Float64:
			acc := accV.Float()
			if accV.Float() >= xV.Float() {
				return acc
			} else {
				return x
			}
		case reflect.String:
			if accV.Float() >= xV.Float() {
				return acc
			} else {
				return x
			}
		}
		return nil
	}, thunk.Head())
}

// Retrieves the minimum element of a list. Obviously, the list must be
// composed of ordered elements (ints, floats or strings).
func (thunk *Thunk) Min() I {
	return thunk.Reduce(func(acc, x I) I {
		accV := reflect.ValueOf(acc)
		xV := reflect.ValueOf(x)
		if accV.Kind() != xV.Kind() {
			return nil
		}
		switch accV.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if accV.Int() <= xV.Int() {
				return acc
			} else {
				return x
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			if accV.Uint() <= xV.Uint() {
				return acc
			} else {
				return x
			}
		case reflect.Float32, reflect.Float64:
			acc := accV.Float()
			if accV.Float() <= xV.Float() {
				return acc
			} else {
				return x
			}
		case reflect.String:
			if accV.Float() <= xV.Float() {
				return acc
			} else {
				return x
			}
		}
		return nil
	}, thunk.Head())
}

// Lists the first elements of the list that pass a filtering function.
func (thunk *Thunk) TakeWhile(f func(I) bool) *Thunk {
	var ret thunkFunc = func(_ I) *Pair {
		pair := force(thunk, nil)
		if pair != nil && f(pair.head) {
			return &Pair{pair.head, pair.tail.TakeWhile(f)}
		}
		return nil
	}
	return &Thunk{nextId(), ret}
}

// Lists the elements of the list after the first one that doesn't pass a 
// filtering function.
func (thunk *Thunk) DropWhile(f func(I) bool) *Thunk {
	var ret thunkFunc
	ret = func(_ I) *Pair {
		pair := force(thunk, nil)
		if pair != nil && f(pair.head) {
			thunk = pair.tail
			return ret(nil)
		}
		return pair
	}
	return &Thunk{nextId(), ret}
}

// Takes some lists and returns a list with slices of one element of each list.
//	ZipN(L(1, 2, 3), L(4, 5, 6)) // L([1 4], [2 5], [3 6])
func ZipN(thunks ...*Thunk) *Thunk {
	return MapN(func(xs ...I) I {
		return SliceToList(xs)
	}, thunks...)
}

// Returns a list with slices of one element of each list.
//	L(1, 2, 3).Zip(L(4, 5, 6)) // L([1 4], [2 5], [3 6])
func (thunk *Thunk) Zip(other *Thunk) *Thunk {
	return ZipN(thunk, other)
}

// Converts a list of lists and makes a single list.
// 	L(L(1, 2), L(3, 4)).Flatten() // L(1, 2, 3, 4)
func (thunk *Thunk) Flatten() *Thunk {
	// Can do better.
	var ret thunkFunc = func(_ I) *Pair {
		pair := force(thunk, nil)
		if pair != nil {
			pair2 := force(pair.head.(*Thunk), nil)
			return &Pair{pair2.head,
				pair2.tail.Append(pair.tail.Flatten())}
		}
		return nil
	}
	return &Thunk{nextId(), ret}
	/*return thunk.Reduce(func(acc, x I) I {
		return acc.(*Thunk).Append(x.(*Thunk))
	}, L()).(*Thunk)*/
}

func (thunk *Thunk) Reverse() *Thunk {
	var ret thunkFunc = func(_ I) *Pair {
		return force(thunk.Reduce(func(acc, x I) I {
			return Link(x, acc.(*Thunk))
		}, L()).(*Thunk), nil)
	}
	return &Thunk{nextId(), ret}
}

func (thunk *Thunk) Last() I {
	pair := force(thunk.Reverse().Take(1), nil)
	return pair.head
}

// Makes an autoupdating infinite list. Each element will be
// generated by a function that takes the previous element as
// argument. You must provide an initial element.
//	naturals = Updating(0, func(x I) I {
//		return x.(int) + 1
//	})
func Updating(initial I, f func(I) I) *Thunk {
	var tail thunkFunc
	tail = func(ctx I) *Pair {
		h := f(ctx)
		var t thunkFunc = func(ctx I) *Pair { return force(&Thunk{nextId(), tail}, h) }
		return &Pair{h, &Thunk{nextId(), t}}
	}
	var ret thunkFunc
	ret = func(_ I) *Pair {
		var t thunkFunc = func(ctx I) *Pair { return force(&Thunk{nextId(), tail}, initial) }
		return &Pair{initial, &Thunk{nextId(), t}}
	}
	return &Thunk{nextId(), ret}
}

func init() {
	memo = true
	memoTable = make(map[uint64]*Pair)
}
