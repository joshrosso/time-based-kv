package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// TimeMap holds a key and all values added over time for that key.
type TimeMap struct {
	times map[string]*timeStore
}

// timeStore is the underlying store for each key.
type timeStore struct {
	timeIndex map[time.Time]*value // mapping to each value based on time
	values    []*value             // underlying data store
}

// value represents the object stored.
type value struct {
	stamp time.Time // timestamp of insertion
	val   string    // string used for simplicity, but imagine a larger struct
}

func main() {
	// NOTE: comment out parts of this to make test output
	// .     easier to read!
	tm := New()

	// validate Set
	kvs := map[string][]string{
		"dog": {"woof", "bark", "sigh", "growl", "wimper"},
		"cat": {"hiss", "screech", "crash", "meow"},
	}
	for k, sounds := range kvs {
		for _, sound := range sounds {
			tm.Set(k, sound)
		}
	}
	spew.Dump(tm)

	// validate Get for latest
	spew.Dump(tm.Get("dog"))

	// validate get for each stamp
	for stamp := range tm.times["dog"].timeIndex {
		noise, err := tm.Get("dog", stamp)
		if err != nil {
			panic(err)
		}
		fmt.Printf("At %s: %s\n", stamp, noise.val)
	}

	// collect a list of all stamps for dog
	collectedStamps := []time.Time{}
	for _, v := range tm.times["dog"].values {
		collectedStamps = append(collectedStamps, v.stamp)
	}
	stampIdxToTest := 2
	fmt.Printf("searching before %s\n", collectedStamps[stampIdxToTest])

	// Uncomment this and pass testTime below to test a time before all elements
	//testTime, err := time.Parse("2006-01-02", "2006-01-02")
	// if err != nil {
	// 	panic(err)
	// }

	before, err := tm.GetBefore("dog", collectedStamps[stampIdxToTest])
	if err != nil {
		panic(err)
	}
	spew.Dump(before)

}

// Set adds a value to a given key. When the value is added, its timestamp is
// recorded.
func (tm *TimeMap) Set(key, data string) {
	// when the key exists, insert the value into the store
	if timeStore, ok := tm.times[key]; ok {
		stamp := time.Now()
		v := &value{stamp: stamp, val: data}
		timeStore.values = append(timeStore.values, v)
		timeStore.timeIndex[stamp] = v
		return
	}
	// when the key is new, create a new store for the key and recall this
	// method
	tm.times[key] = newTimeStore()
	tm.Set(key, data)
}

// Get returns a value for key. If stamp is provided, the value with that
// timestamp is returned. If stamp is not provided, the last element inserted
// under key is returned.
func (tm *TimeMap) Get(key string, stamp ...time.Time) (*value, error) {
	ts, err := tm.getTimeStore(key)
	if err != nil {
		return nil, err
	}

	// return latest
	if len(stamp) < 1 {
		return ts.values[len(ts.values)-1], nil
	}
	lookup := stamp[0]
	if v, ok := ts.timeIndex[lookup]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("key [%s] had no timestamp [%s]", key, lookup)
}

// GetBefore returns all values stored for key, where their insertion time is
// before stamp.
func (tm *TimeMap) GetBefore(key string, stamp time.Time) ([]*value, error) {
	ts, err := tm.getTimeStore(key)
	if err != nil {
		return nil, err
	}

	// locate the lowest element with a timestamp lower than stamp
	idx := sort.Search(len(ts.values), func(i int) bool {
		return ts.values[i].stamp.After(stamp)
	})
	// return the list in for of [0:n). When n is 0, meaning the first element
	// is after stamp, an empty list is returned.
	return ts.values[:idx], nil
}

// New returns a new [TimeMap].
func New() TimeMap {
	return TimeMap{
		times: map[string]*timeStore{},
	}
}

// newTimeStore is used when a new key is introduced. It intializes and returns
// the pointer to the new key's store.
func newTimeStore() *timeStore {
	return &timeStore{
		timeIndex: map[time.Time]*value{},
		values:    []*value{},
	}
}

func (tm *TimeMap) getTimeStore(key string) (*timeStore, error) {
	if ts, ok := tm.times[key]; ok {
		return ts, nil
	}
	return nil, fmt.Errorf("key [%s] does not exist", key)
}
