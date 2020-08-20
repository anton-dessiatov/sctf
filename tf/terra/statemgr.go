package terra

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/anton-dessiatov/sctf/tf/dal"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/jinzhu/gorm"
)

func (t *Terra) NewStateMgr(id StackIdentity) (statemgr.Full, error) {
	return newDBStateMgr(t.db, id)
}

type dbStateMgr struct {
	db      *gorm.DB
	stackID int

	current *states.State
}

func newDBStateMgr(db *gorm.DB, id StackIdentity) (*dbStateMgr, error) {
	var dalStack dal.Stack
	err := db.First(&dalStack).Where("cluster_id = ? AND name = ?", id.ClusterID, id.Name).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		dalStack, err = createDBStack(db, id)
		if err != nil {
			return nil, fmt.Errorf("createStack: %w", err)
		}
	case err != nil:
		return nil, fmt.Errorf("st.db.First: %w", err)
	default:
		// All okay, keep up
	}

	return &dbStateMgr{
		db:      db,
		stackID: dalStack.ID,
	}, nil
}

func createDBStack(db *gorm.DB, id StackIdentity) (dal.Stack, error) {
	result := dal.Stack{
		ClusterID: id.ClusterID,
		Name:      id.Name,
	}

	err := db.Create(&result).Error
	if err != nil {
		return result, fmt.Errorf("db.Create: %w", err)
	}

	return result, nil
}

// statemgr.Locker

func (st *dbStateMgr) Lock(info *statemgr.LockInfo) (string, error) {
	// TODO: it's dangerous to not have the lock in distributed scenario. We need to implement it.
	// For the prototype it's fine, however
	return "", nil
}

func (st *dbStateMgr) Unlock(id string) error {
	// TODO: it's dangerous to not have the lock in distributed scenario. We need to implement it.
	// For the prototype it's fine, however
	return nil
}

// statemgr.Transient

// State returns the latest state.
//
// Each call to State returns an entirely-distinct copy of the state, with
// no storage shared with any other call, so the caller may freely mutate
// the returned object via the state APIs.
func (st *dbStateMgr) State() *states.State {
	if st.current == nil {
		st.current = states.NewState()
	}
	return st.current.DeepCopy()
}

func (st *dbStateMgr) commit() error {
	f := statefile.New(st.current, "sctf", 0)
	var b bytes.Buffer
	err := statefile.Write(f, &b)
	if err != nil {
		return fmt.Errorf("statefile.Write: %w", err)
	}

	dalState := dal.State{
		StackID: st.stackID,
		Body:    b.Bytes(),
	}

	err = st.db.Create(&dalState).Error
	if err != nil {
		return fmt.Errorf("st.db.Create: %w", err)
	}

	return nil
}

// Write state saves a transient snapshot of the given state.
//
// The caller must ensure that the given state object is not concurrently
// modified while a WriteState call is in progress. WriteState itself
// will never modify the given state.
func (st *dbStateMgr) WriteState(newSt *states.State) error {
	st.current = newSt
	if err := st.commit(); err != nil {
		return fmt.Errorf("st.commit: %w", err)
	}
	return nil
}

// RefreshState retrieves a snapshot of state from persistent storage,
// returning an error if this is not possible.
//
// Types that implement RefreshState generally also implement a State
// method that returns the result of the latest successful refresh.
func (st *dbStateMgr) RefreshState() error {
	var dbState dal.State
	err := st.db.Where("stack_id = ?", st.stackID).Order("id desc").First(&dbState).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// That's okay, there is no state in the database yet, but it's not a huge deal
		return nil
	case err != nil:
		return fmt.Errorf("st.db.First: %w", err)
	default:
		// Go on. We've got ourselves JSON
	}

	// Parse statefile from body
	f, err := statefile.Read(bytes.NewReader(dbState.Body))
	if err != nil {
		return fmt.Errorf("statefile.Read: %w", err)
	}

	st.current = f.State
	return nil
}

func (st *dbStateMgr) PersistState() error {
	return nil
}
