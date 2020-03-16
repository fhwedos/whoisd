package memcache

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fhwedos/whoisd/pkg/config"
	"github.com/patrickmn/go-cache"
)

// Record - standard record (struct) for config package
type Record struct {
	config *config.Record
	C      *cache.Cache
}

// CacheControl - cache control commands
type CacheControl struct {
	flushCache bool
	listCache  bool
}

// simplest logger, which initialized during starts of the application
var (
	stdlog = log.New(os.Stdout, "[CACHE]: ", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "[CACHE:ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// New - returns new config record initialized with default values
func New(conf *config.Record) (*Record, error) {
	if conf.CacheEnabled != true {
		return nil, errors.New("Cache is not enabled")
	}

	c := cache.New(
		time.Duration(conf.CacheExpiration)*time.Minute,
		time.Duration(conf.CacheCleanupInterval)*time.Minute,
	)

	return &Record{conf, c}, nil
}

// checkCacheControl - checks and proccess cache control commands
func (memcache *Record) checkCacheControl() error {
	_, err := os.Stat("/etc/whoisd/cache.control")
	if err != nil {
		return err
	}

	err = os.Remove("/etc/whoisd/cache.control")
	if err != nil {
		errlog.Println("Failed to remove cache.control file.")
	}

	control := getCacheControl(memcache.config.CacheControlPath)

	if control.listCache == true {
		stdlog.Println("Cache list")
	}

	if control.flushCache == true {
		stdlog.Printf("Flush cache. %d items removed from cache.", memcache.C.ItemCount())
		memcache.C.Flush()
	}

	return nil
}

// initCacheControl - init cache control configuration
func getCacheControl(path string) *CacheControl {
	control := new(CacheControl)
	flag.BoolVar(&control.flushCache, "flush", false, "removes all items from cache")
	flag.BoolVar(&control.listCache, "list", false, "list all items in cache into file")
	loadCacheControlFile(control, path)
	return control
}

// loadCacheControlFile - loads cache control file into CacheControl record
func loadCacheControlFile(control *CacheControl, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(bufio.NewReader(file)).Decode(&control); err != nil {
		return err
	}

	return nil
}

// WriteCacheControl - writes cache control configuration
func WriteCacheControl(path string, flush bool, list bool) error {
	control := new(CacheControl)
	control.flushCache = flush
	control.listCache = list

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		errlog.Println("Failed to open cache control file.")
		return err
	}

	if err := json.NewEncoder(bufio.NewWriter(file)).Encode(&control); err != nil {
		errlog.Println("Failed to write cache control configuration")
		return err
	}

	defer file.Close()

	ctrlFile, err := os.OpenFile("/etc/whoisd/cache.control", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		errlog.Println("Failed to create cache control file")
		return err
	}

	defer ctrlFile.Close()

	return nil
}

// Set - save item in cache
func (memcache *Record) Set(key string, value interface{}) {
	memcache.C.Set(key, value, cache.DefaultExpiration)
	stdlog.Println("Items cached: ", memcache.C.ItemCount())
}

// Get - get item from cache
func (memcache *Record) Get(key string) (interface{}, bool) {
	memcache.checkCacheControl()
	return memcache.C.Get(key)
}
