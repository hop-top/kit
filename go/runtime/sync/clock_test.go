package sync

import (
	"sync"
	"testing"
)

func TestClock_Now_Monotonic(t *testing.T) {
	c := NewClock("node1")
	prev := c.Now()
	for i := 0; i < 100; i++ {
		cur := c.Now()
		if !prev.Before(cur) {
			t.Fatalf("iteration %d: %+v not before %+v", i, prev, cur)
		}
		prev = cur
	}
}

func TestClock_Now_SetsNodeID(t *testing.T) {
	c := NewClock("abc")
	ts := c.Now()
	if ts.NodeID != "abc" {
		t.Fatalf("expected NodeID abc, got %s", ts.NodeID)
	}
}

func TestClock_Update_TakesMax(t *testing.T) {
	c := NewClock("local")
	local := c.Now()

	remote := Timestamp{
		Physical: local.Physical + 1_000_000_000,
		Logical:  5,
		NodeID:   "remote",
	}

	merged := c.Update(remote)
	if merged.Physical < remote.Physical {
		t.Fatalf("merged physical %d < remote %d", merged.Physical, remote.Physical)
	}
	if !local.Before(merged) {
		t.Fatalf("local %+v not before merged %+v", local, merged)
	}
}

func TestTimestamp_Before_Ordering(t *testing.T) {
	a := Timestamp{Physical: 1, Logical: 0, NodeID: "a"}
	b := Timestamp{Physical: 2, Logical: 0, NodeID: "a"}
	if !a.Before(b) {
		t.Fatal("expected a before b by physical")
	}

	c := Timestamp{Physical: 1, Logical: 1, NodeID: "a"}
	if !a.Before(c) {
		t.Fatal("expected a before c by logical")
	}

	d := Timestamp{Physical: 1, Logical: 0, NodeID: "b"}
	if !a.Before(d) {
		t.Fatal("expected a before d by NodeID")
	}
}

func TestTimestamp_Equal(t *testing.T) {
	a := Timestamp{Physical: 10, Logical: 3, NodeID: "x"}
	b := Timestamp{Physical: 10, Logical: 3, NodeID: "x"}
	if !a.Equal(b) {
		t.Fatal("expected equal")
	}
	b.Logical = 4
	if a.Equal(b) {
		t.Fatal("expected not equal")
	}
}

func TestClock_ConcurrentSafety(t *testing.T) {
	c := NewClock("node")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Now()
			}
		}()
	}
	wg.Wait()
}
