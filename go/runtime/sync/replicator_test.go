package sync

import (
	"context"
	"fmt"
	gosync "sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/kit/go/runtime/domain"
)

type replTestEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (e replTestEntity) GetID() string { return e.ID }

type replMemRepo struct {
	mu   gosync.RWMutex
	data map[string]*replTestEntity
}

func newReplMemRepo() *replMemRepo {
	return &replMemRepo{data: make(map[string]*replTestEntity)}
}

func (r *replMemRepo) Create(_ context.Context, e *replTestEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[e.GetID()]; ok {
		return domain.ErrConflict
	}
	r.data[e.GetID()] = e
	return nil
}
func (r *replMemRepo) Get(_ context.Context, id string) (*replTestEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return e, nil
}
func (r *replMemRepo) List(_ context.Context, _ domain.Query) ([]replTestEntity, error) {
	return nil, nil
}
func (r *replMemRepo) Update(_ context.Context, e *replTestEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[e.GetID()]; !ok {
		return domain.ErrNotFound
	}
	r.data[e.GetID()] = e
	return nil
}
func (r *replMemRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func TestReplicator_SingleRemotePush(t *testing.T) {
	repo := newReplMemRepo()
	mt := NewMemoryTransport()
	rem := Remote{Name: "origin", Transport: mt, Mode: PushOnly}

	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](rem),
		WithInterval[replTestEntity](50*time.Millisecond),
	)

	ts := Timestamp{Physical: 100, Logical: 0, NodeID: "local"}
	rep.Enqueue(Diff{EntityID: "e1", Operation: OpCreate, Timestamp: ts, After: []byte(`{"id":"e1","name":"test"}`)})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, rep.Start(ctx))
	defer rep.Stop()

	time.Sleep(100 * time.Millisecond)

	diffs := mt.Diffs()
	assert.Len(t, diffs, 1)
	assert.Equal(t, "e1", diffs[0].EntityID)
}

func TestReplicator_SingleRemotePull(t *testing.T) {
	repo := newReplMemRepo()
	mt := NewMemoryTransport()

	// Pre-populate remote with data
	_ = mt.Push(context.Background(), []Diff{
		{EntityID: "e1", Operation: OpCreate, Timestamp: Timestamp{Physical: 50, NodeID: "remote"}, After: []byte(`{"id":"e1","name":"pulled"}`)},
	})

	rem := Remote{Name: "origin", Transport: mt, Mode: PullOnly}
	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](rem),
		WithInterval[replTestEntity](50*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, rep.Start(ctx))
	defer rep.Stop()

	time.Sleep(100 * time.Millisecond)

	// Entity should be created in repo
	e, err := repo.Get(ctx, "e1")
	require.NoError(t, err)
	assert.Equal(t, "pulled", e.Name)
}

func TestReplicator_MultiRemote(t *testing.T) {
	repo := newReplMemRepo()
	mt1 := NewMemoryTransport()
	mt2 := NewMemoryTransport()

	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](Remote{Name: "r1", Transport: mt1, Mode: PushOnly}),
		WithRemote[replTestEntity](Remote{Name: "r2", Transport: mt2, Mode: PushOnly}),
		WithInterval[replTestEntity](50*time.Millisecond),
	)

	rep.Enqueue(Diff{EntityID: "e1", Operation: OpCreate, Timestamp: Timestamp{Physical: 1, NodeID: "local"}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, rep.Start(ctx))
	defer rep.Stop()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, mt1.Diffs(), 1)
	assert.Len(t, mt2.Diffs(), 1)
}

func TestReplicator_FilterApplied(t *testing.T) {
	repo := newReplMemRepo()
	mt := NewMemoryTransport()

	// Filter only allows entity "e2"
	rem := Remote{
		Name:      "filtered",
		Transport: mt,
		Mode:      PushOnly,
		Filter:    func(d Diff) bool { return d.EntityID == "e2" },
	}

	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](rem),
		WithInterval[replTestEntity](50*time.Millisecond),
	)

	rep.Enqueue(Diff{EntityID: "e1", Operation: OpCreate, Timestamp: Timestamp{Physical: 1, NodeID: "local"}})
	rep.Enqueue(Diff{EntityID: "e2", Operation: OpCreate, Timestamp: Timestamp{Physical: 2, NodeID: "local"}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, rep.Start(ctx))
	defer rep.Stop()

	time.Sleep(100 * time.Millisecond)

	diffs := mt.Diffs()
	assert.Len(t, diffs, 1)
	assert.Equal(t, "e2", diffs[0].EntityID)
}

func TestReplicator_QueueBounds(t *testing.T) {
	repo := newReplMemRepo()
	mt := NewMemoryTransport()
	rem := Remote{Name: "origin", Transport: mt, Mode: PushOnly}

	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](rem),
		WithInterval[replTestEntity](1*time.Hour), // don't tick during test
	)

	// Enqueue more than maxQueueSize items.
	for i := range 10100 {
		rep.Enqueue(Diff{
			EntityID:  fmt.Sprintf("e%d", i),
			Operation: OpCreate,
			Timestamp: Timestamp{Physical: int64(i), NodeID: "local"},
		})
	}

	// Queue should never exceed maxQueueSize (10000).
	rep.mu.RLock()
	qLen := len(rep.queue)
	rep.mu.RUnlock()
	assert.LessOrEqual(t, qLen, 10000, "queue exceeded maxQueueSize")
	assert.Equal(t, 10000, qLen)
}

func TestReplicator_ModeRespected(t *testing.T) {
	repo := newReplMemRepo()
	mt := NewMemoryTransport()

	// PullOnly should not push
	rem := Remote{Name: "pull-only", Transport: mt, Mode: PullOnly}
	rep := NewReplicator[replTestEntity](repo,
		WithRemote[replTestEntity](rem),
		WithInterval[replTestEntity](50*time.Millisecond),
	)

	rep.Enqueue(Diff{EntityID: "e1", Operation: OpCreate, Timestamp: Timestamp{Physical: 1, NodeID: "local"}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, rep.Start(ctx))
	defer rep.Stop()

	time.Sleep(100 * time.Millisecond)

	// Nothing should be pushed
	assert.Empty(t, mt.Diffs())
}
