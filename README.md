# Interval Tree
An interval tree is a data structure that stores elements, and allows for efficient range queries. 

This is a generic, goroutine-safe implementation, with configurable collision strategies.

# Examples

## `Unique` [int, string]
```
it := intervaltree.Unique[int, string](func(a, b int) bool {
    return a < b
})

// add some elements
it.Insert(1, "A")
it.Insert(2, "B")
it.Insert(3, "C")
it.Insert(4, "D")
it.Insert(5, "E")

// `Unique` strategy overrides duplicates
it.Insert(3, "X")

// remove 4
it.Delete(4)

res := it.Search(2, 5)
// []intervaltree.Entry[int, string]{{2, "B"}, {3, "X"}, {5, "E"}}

```

## `Duplicates` [string, int]

```
it2 := intervaltree.Duplicates[string, int](func(a, b string) bool {
    return a < b
})

// add some elements
it2.Insert("A", 1)
it2.Insert("B", 2)
it2.Insert("C", 3)
it2.Insert("D", 4)

// duplicates are allowed
it2.Insert("B", 5)

result2 := it2.Search("A", "C")
/*
[]Entry[string, int]{
    {Key: "A", Value: 1},
    {Key: "B", Value: 2},
    {Key: "B", Value: 5},
    {Key: "C", Value: 3},
})
*/