package site

import (
	"dna"
	"dna/item"
	"dna/sqlpg"
	"sync"
)

// Range defines a range from first element to last one
type Range struct {
	First dna.Int
	Last  dna.Int
	Total dna.Int
}

func NewRange(first, last dna.Int) *Range {
	total := last - first + 1
	return &Range{first, last, total}
}

func (r Range) String() string {
	return dna.Sprintf("[%v=>%v] - %v", r.First, r.Last, r.Total).String()
}

// StateHandler defines the state of Update(). It ensures its fields are only called once.
//
// StateHandler resolves 3 common patterns to update new item from a site.
//
// 	* Pattern 1: Update items from last ids of (song, album...) of a site to the newest ones.
// 	It will stop after N continuous failures which comes from SiteConfig. It means the newest ones are found.
// 	The Cid is the lasted id of item in a table.
// 	Fields used: IsOver
//
// 	* Pattern 2: Update through range. Usually to fetch items from m to n.
// 	Ex: getting all songs from X with range from 1000 to 2000.
// 	Fields used: Cid, IsOver, Range
//
// 	* Pattern 3: Update through an external slice.
// 	Ex: Re-getting all failed ids from log file whose ids are not in order.
// 	Or in the case of nhaccuatui, a key is encrypted and is used to display a page, so songid is hidden.
// 	Therefore no way to loop through a range. Only list of keys is found, which is translated into ids.
// 	The ids is an integer slice. So 3rd pattern will be applied.
// 	Fields used: Cid, IsOver, ExtSlice
type StateHandler struct {
	mu         sync.RWMutex
	Cid        dna.Int       // Current ID of a page Update() is getting.
	SiteConfig *SiteConfig   // Site config containing n continuous failures - NCFail (1st pattern)
	Range      *Range        // Looping through range if available (2nd pattern)
	ExtSlice   *dna.IntArray // Looping through an slice (3rd pattern)
	Db         *sqlpg.DB     // Connection-to-db state
	Pattern    dna.Int       // pattern number
	IsOver     dna.Bool      // IsOver is true when nothing to be updated
	Item       item.Item
	TableName  dna.String
	NcCount    dna.Int // The number of concurrency count
}

// CheckStateHandler panics if StateHandler is not in proper format.
func CheckStateHandler(state *StateHandler) {
	switch state.GetPattern() {
	case 1:
		if state.GetRange() != nil || state.GetExtSlice() != nil {
			panic("Wrong 1st pattern!")
		}
	case 2:
		if state.GetRange() == nil || state.GetExtSlice() != nil {
			panic("Wrong 2nd pattern")
		}
	case 3:
		if state.GetExtSlice() == nil || state.GetRange() != nil {
			panic("Wrong 3rd pattern")
		}
	default:
		panic("Wrong pattern number")
	}
}

// NewStateHandler returns default updates (1st pattern).
func NewStateHandler(itm item.Item, config *SiteConfig, db *sqlpg.DB) *StateHandler {
	tableName := sqlpg.GetTableName(itm)
	return &StateHandler{
		Cid:        0,
		SiteConfig: config,
		Range:      nil,
		ExtSlice:   nil,
		Db:         db,
		Pattern:    1,
		IsOver:     false,
		Item:       itm,
		TableName:  tableName,
		NcCount:    0,
	}
}

// NewStateHandlerWithRange returns new StateHandler with 2nd pattern.
func NewStateHandlerWithRange(itm item.Item, r *Range, config *SiteConfig, db *sqlpg.DB) *StateHandler {
	tableName := sqlpg.GetTableName(itm)
	return &StateHandler{
		Cid:        r.First - 1,
		SiteConfig: config,
		Range:      r,
		ExtSlice:   nil,
		Db:         db,
		Pattern:    2,
		IsOver:     false,
		Item:       itm,
		TableName:  tableName,
		NcCount:    0,
	}
}

// NewStateHandlerWithExtSlice returns  new StateHandler with 3rd pattern.
func NewStateHandlerWithExtSlice(itm item.Item, extSlice *dna.IntArray, config *SiteConfig, db *sqlpg.DB) *StateHandler {
	tableName := sqlpg.GetTableName(itm)
	return &StateHandler{
		Cid:        -1, // In this case, Cid means index of current element in external slice
		SiteConfig: config,
		Range:      nil,
		ExtSlice:   extSlice,
		Db:         db,
		Pattern:    3,
		IsOver:     false,
		Item:       itm,
		TableName:  tableName,
		NcCount:    0,
	}
}

// IncreaseCid increases Cid by a unit and returns the increased id.
func (sh *StateHandler) IncreaseCid() {
	sh.mu.Lock()
	sh.Cid += 1
	sh.mu.Unlock()
}

func (sh *StateHandler) SetCid(n dna.Int) {
	sh.mu.Lock()
	sh.Cid = n
	sh.mu.Unlock()
}

// GetCid returns Cid.
func (sh *StateHandler) GetCid() dna.Int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if sh.Pattern == 1 || sh.Pattern == 2 {
		return sh.Cid
	} else {
		idx := sh.Cid
		length := sh.ExtSlice.Length()
		if idx >= length {
			idx = length - 1
		}
		return (*sh.ExtSlice)[idx]
	}
}

// IsComplete returns the value of IsOver
func (sh *StateHandler) IsComplete() dna.Bool {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.IsOver
}

// SetCompletion sets IsOver to be true
func (sh *StateHandler) SetCompletion() {
	sh.mu.Lock()
	sh.IsOver = true
	sh.mu.Unlock()
}

// AddNcCount adds n to NcCount
func (sh *StateHandler) AddNcCount(n dna.Int) {
	sh.mu.Lock()
	sh.NcCount += n
	sh.mu.Unlock()
}

// AddNcCount adds n to NcCount
func (sh *StateHandler) GetNcCount() dna.Int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.NcCount
}

func (sh *StateHandler) GetRange() *Range {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.Range
}

func (sh *StateHandler) GetPattern() dna.Int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.Pattern
}

func (sh *StateHandler) GetExtSlice() *dna.IntArray {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.ExtSlice
}

func (sh *StateHandler) GetNCFail() dna.Int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	switch {
	case sh.TableName.Match(`song`) == true:
		return sh.SiteConfig.NCSongFail
	case sh.TableName.Match(`album`) == true:
		return sh.SiteConfig.NCAlbumFail
	case sh.TableName.Match(`video`) == true:
		return sh.SiteConfig.NCVideoFail
	default:
		// WARNING.Println("Cannot find type of NCFail: it has to be song, album or video")
		// WARNING.Println("It returns default value.")
		return sh.SiteConfig.NCSongFail
	}
}

func (sh *StateHandler) GetItem() item.Item {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.Item
}

func (sh *StateHandler) GetTableName() dna.String {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.TableName
}

func (sh *StateHandler) GetDb() *sqlpg.DB {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.Db
}

// InsertIgnore implements StateHandler.Db.InsertIgnore() method
// func (sh *StateHandler) InsertIgnore(itm item.Item) {
// 	go func() {
// 		err := sh.Db.InsertIgnore(itm)
// 		if err != nil {
// 			dna.Log("Cannot insert new item. ", err.Error())
// 		}
// 	}()
// }
