# TODO

- How should nil in a key-value be handled? Currently, a nil
  Value will crash. How can this be made safer?

- The interaction between the tuple iterator & compare is
  difficult to understand. I don't think the iterator
  abstraction is very useful. It could be removed and the
  compare package simplified.

- Should errors be defined at the top of all packages?
  Should they be made public as part of the interface? If
  so, a custom error implementation should be built to
  facilitate this. See the engine package for an initial
  example.

