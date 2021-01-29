/*
Open Source Initiative OSI - The MIT License (MIT):Licensing

The MIT License (MIT)
Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mapset

import "sync"

type threadSafeSet struct {
	objects threadUnsafeSet
	mutex   sync.RWMutex
}

func newThreadSafeSet() threadSafeSet {
	return threadSafeSet{objects: newThreadUnsafeSet()}
}

func (set *threadSafeSet) Add(i interface{}) bool {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	return set.objects.Add(i)
}

func (set *threadSafeSet) Contains(i ...interface{}) bool {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return set.objects.Contains(i...)
}

func (set *threadSafeSet) IsSubset(other Set) bool {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return set.objects.IsSubset(&o.objects)
}

func (set *threadSafeSet) IsProperSubset(other Set) bool {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return set.objects.IsProperSubset(&o.objects)
}

func (set *threadSafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *threadSafeSet) IsProperSuperset(other Set) bool {
	return other.IsProperSubset(set)
}

func (set *threadSafeSet) Union(other Set) Set {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	union := set.objects.Union(&o.objects).(*threadUnsafeSet)
	return &threadSafeSet{objects: *union}
}

func (set *threadSafeSet) Intersect(other Set) Set {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	unsafeIntersection := set.objects.Intersect(&o.objects).(*threadUnsafeSet)
	return &threadSafeSet{objects: *unsafeIntersection}
}

func (set *threadSafeSet) Difference(other Set) Set {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	diff := set.objects.Difference(&o.objects).(*threadUnsafeSet)
	return &threadSafeSet{objects: *diff}
}

func (set *threadSafeSet) SymmetricDifference(other Set) Set {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	diff := set.objects.SymmetricDifference(&o.objects).(*threadUnsafeSet)
	return &threadSafeSet{objects: *diff}
}

func (set *threadSafeSet) Clear() {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	set.objects = newThreadUnsafeSet()
}

func (set *threadSafeSet) Remove(i interface{}) {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	delete(set.objects, i)
}

func (set *threadSafeSet) Cardinality() int {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return len(set.objects)
}

func (set *threadSafeSet) Length() int {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return len(set.objects)
}

func (set *threadSafeSet) Each(callback func(interface{}) bool) {
	set.mutex.RLock()
	defer set.mutex.RUnlock()
	for elem := range set.objects {
		if callback(elem) {
			break
		}
	}
}

func (set *threadSafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		set.mutex.RLock()
		for elem := range set.objects {
			ch <- elem
		}
		close(ch)
		set.mutex.RUnlock()
	}()

	return ch
}

func (set *threadSafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
		set.mutex.RLock()
	L:
		for elem := range set.objects {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
		set.mutex.RUnlock()
	}()

	return iterator
}

func (set *threadSafeSet) Equal(other Set) bool {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return set.objects.Equal(&o.objects)
}

func (set *threadSafeSet) Clone() Set {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	clone := set.objects.Clone().(*threadUnsafeSet)
	return &threadSafeSet{objects: *clone}
}

func (set *threadSafeSet) String() string {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return set.objects.String()
}

func (set *threadSafeSet) PowerSet() Set {
	set.mutex.RLock()
	unsafePowerSet := set.objects.PowerSet().(*threadUnsafeSet)
	set.mutex.RUnlock()

	tss := &threadSafeSet{objects: newThreadUnsafeSet()}
	for subset := range unsafePowerSet.Iter() {
		unsafeSubset := subset.(*threadUnsafeSet)
		tss.Add(&threadSafeSet{objects: *unsafeSubset})
	}

	return tss
}

func (set *threadSafeSet) Pop() interface{} {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	return set.objects.Pop()
}

func (set *threadSafeSet) CartesianProduct(other Set) Set {
	o := other.(*threadSafeSet)

	set.mutex.RLock()
	defer set.mutex.RUnlock()
	o.mutex.RLock()
	defer set.mutex.RUnlock()

	// unsafe cartesian product
	ucp := set.objects.CartesianProduct(&o.objects).(*threadUnsafeSet)
	return &threadSafeSet{objects: *ucp}
}

func (set *threadSafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, set.Cardinality())

	set.mutex.RLock()
	defer set.mutex.RUnlock()

	for elem := range set.objects {
		keys = append(keys, elem)
	}

	return keys
}

func (set *threadSafeSet) Strings() []string {
	keys := make([]string, 0, set.Length())

	set.mutex.RLock()
	defer set.mutex.RUnlock()

	for elem := range set.objects {
		switch elem.(type) {
		case string:
			keys = append(keys, elem.(string))
		}
	}

	return keys
}

func (set *threadSafeSet) MarshalJSON() ([]byte, error) {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return set.objects.MarshalJSON()
}

func (set *threadSafeSet) UnmarshalJSON(p []byte) error {
	set.mutex.RLock()
	defer set.mutex.RUnlock()

	return set.objects.UnmarshalJSON(p)
}
