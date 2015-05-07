//This file is responsible for interfaces and structs needed for Caching

package cache


// Create a new Cache.
func NewCache(maxEntries int) *Cache {
	return &Cache{
	}
}

//Add an new key value pair to the map
func (c *Cache) Add(key int, value) {} 

//Lookup based on id
func (c *Cache) LookupId(id int) (api.StateInfo) {}

//Find the entry to evict
func (c *Cache) LRU() int{}


