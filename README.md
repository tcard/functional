# functional

A functional programming library including
a lazy list implementation and some of the most usual functions.

	import FP "github.com/tcard/functional"
	
## Installing

	go get github.com/tcard/functional

Tested on Go version weekly.2012-03-22.
	
## Examples

The Fibonacci sequence, infinite stream:

	var fibo *FP.Thunk
	fibo = FP.Link(1, FP.DelayedLink(1, func() *FP.Thunk {
		return FP.MapN(func(xs ...FP.I) FP.I {
			ret := 0
			for _, v := range xs {
				ret += v.(int)
			}
			return ret
		}, fibo, fibo.Tail())
	}))
	
Which is an uglier, slower version of the famous Lisp-like:

	(define fibo
		(cons 1
			(cons 1
				(map + fibo (cdr fibo)))))		

For showing some "real" utility, this would retrieve the third page of a blog:

	var posts *FP.Thunk	= ...
	page := 3
	postsPerPage := 10
	_ = posts.Drop((page - 1) * postsPerPage).Take(postsPerPage).Map(func (post FP.I) FP.I {
		blog.showPost(post.(Post))
		return post
	})
	
See documentation and tests for detailed usage.

## To Do
	
* Better memoization.
* Ability to skip lists and work directly with slices, although lack of generics would make it clumsy anyway.
* Some optimizations.