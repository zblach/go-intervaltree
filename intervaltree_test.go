package intervaltree

import (
	"testing"
	"time"
)

func TestIntervalTree(t *testing.T) {
	// Create a new IntervalTree
	tree := New[int, string](func(a, b int) bool {
		return a < b
	}, Replace[int, string])

	// Insert intervals
	tree.Insert(1, "A")
	tree.Insert(3, "B")
	tree.Insert(5, "C")
	tree.Insert(7, "D")

	// Search for overlapping intervals
	result := tree.Search(1, 5)

	// Verify the search results
	assertEqualValues(t, []Entry[int, string]{
		{Key: 1, Value: "A"},
		{Key: 3, Value: "B"},
		{Key: 5, Value: "C"},
	}, result)
}

func TestStringKeys(t *testing.T) {
	// Create a new IntervalTree with string keys
	tree := Duplicates[string, int](func(a, b string) bool {
		return a < b
	})

	// Insert intervals
	tree.Insert("A", 1)
	tree.Insert("B", 2)
	tree.Insert("C", 3)
	tree.Insert("D", 4)

	tree.Insert("B", 5)

	// Search for overlapping intervals
	result := tree.Search("A", "C")

	// Verify the search results
	assertEqualValues(t, []Entry[string, int]{
		{Key: "A", Value: 1},
		{Key: "B", Value: 2},
		{Key: "B", Value: 5},
		{Key: "C", Value: 3},
	}, result)
}

func TestIntervalTreeWithDatetime(t *testing.T) {
	// Create a new IntervalTree with datetime as the key
	tree := Unique[time.Time, string](func(a, b time.Time) bool {
		return a.Before(b)
	})

	// Define some datetime intervals
	start := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, time.December, 31, 23, 59, 59, 999, time.UTC)

	// Insert intervals
	tree.Insert(start, "A")
	tree.Insert(start.Add(time.Hour), "B")
	tree.Insert(end, "C")

	// Search for overlapping intervals
	result := tree.Search(start.Add(-time.Minute), end.Add(-time.Minute))

	// Verify the search results
	assertEqualValues(t, []Entry[time.Time, string]{
		{Key: start, Value: "A"},
		{Key: start.Add(time.Hour), Value: "B"},
	}, result)
}

func TestIntervalTreeWithDuplicates(t *testing.T) {
	// Create a new IntervalTree
	tree := Duplicates[int, string](func(a, b int) bool {
		return a < b
	})

	// Insert intervals
	tree.Insert(1, "A")
	tree.Insert(3, "B")
	tree.Insert(5, "C")
	tree.Insert(7, "D")
	tree.Insert(1, "E")
	tree.Insert(3, "F")
	tree.Insert(5, "G")
	tree.Insert(7, "H")

	tree.Insert(2, "I")

	// Search for overlapping intervals
	result := tree.Search(1, 5)

	// Verify the search results
	assertEqualValues(t, []Entry[int, string]{
		{Key: 1, Value: "A"},
		{Key: 1, Value: "E"},
		{Key: 2, Value: "I"},
		{Key: 3, Value: "B"},
		{Key: 3, Value: "F"},
		{Key: 5, Value: "C"},
		{Key: 5, Value: "G"},
	}, result)
}

func TestReverseOrdering(t *testing.T) {
	// Create a new IntervalTree
	tree := New[int, string](func(a, b int) bool {
		return a > b
	}, Replace[int, string])

	// Insert intervals
	tree.Insert(7, "D")
	tree.Insert(5, "C")
	tree.Insert(3, "B")
	tree.Insert(1, "A")

	// Search for overlapping intervals
	result := tree.Search(3, 8)

	// Verify the search results
	assertEqualValues(t, []Entry[int, string]{
		{Key: 7, Value: "D"},
		{Key: 5, Value: "C"},
		{Key: 3, Value: "B"},
	}, result)
}

func TestDeletion(t *testing.T) {
	// Create a new IntervalTree
	tree := New[int, string](func(a, b int) bool {
		return a < b
	}, Replace[int, string])

	// Insert intervals
	for i := 10; i > 0; i-- {
		tree.Insert(i, "node")
	}

	// Delete an interval
	tree.Delete(3)

	// Search for overlapping intervals
	result := tree.Search(1, 5)

	// Verify the search results
	assertEqualValues(t, []Entry[int, string]{
		{Key: 1, Value: "node"},
		{Key: 2, Value: "node"},
		{Key: 4, Value: "node"},
		{Key: 5, Value: "node"},
	}, result)
}

func assertEqualValues[T, V comparable](t *testing.T, a, b []Entry[T, V]) {
	if len(a) != len(b) {
		t.Fatalf("expected %d entries, got %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Key != b[i].Key {
			t.Fatalf("expected key %v, got %v", a[i].Key, b[i].Key)
		}
		if a[i].Value != b[i].Value {
			t.Fatalf("expected value %v, got %v", a[i].Value, b[i].Value)
		}
	}
}
