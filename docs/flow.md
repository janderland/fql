# Flow Control

FQL will eventually include flow control via scoping.
Certain queries will be guarded by conditional scopes which
control when they are executed.

```fql
$ErrStr=str
$BlobHash=bytes

@cryto/hash(:blob)=<$BlobHash|$ErrStr>

{ <result:$ErrStr>
    @file("err.txt","wa")=:result
}
```

As the example above hints at, there are a couple things to
work out regarding this functionality:

1. I'd like to define custom types using the `$` symbol.
   These custom types would be functionally equivalent to
   data elements but pattern matching at scopes would see
   them as different.

2. Custom types sorta overlap with references in their
   functionality: they are use to direct the output of
   a query to the input of a scope. Is there a way we can
   unify these so that elements are piped to both scopes and
   queries in the same way?

3. We could allow data elements in a variable's type list.
   This would allow us to pattern match a scope on
   a particular value returned by a query. This would be
   similar to how we allow `nil` in the type list, which is
   both a type and a value.

4. It doesn't look like we need type decomposition because
   we can define the components of a compound type as custom
   types and them compose them into a variable during
   querying. In the example above, this is how we only
   execute a scope when the `$ErrStr` sub-type is returned
   by the query.

5. We probably want some basic conditional expressions.
   Things like "less than", "equals", "in this set"?
   Alternatively we can lean into the fact that scopes are
   mostly used for routing data rather than working with it
   in detail.
