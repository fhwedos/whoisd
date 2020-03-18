package memcache

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fhwedos/whoisd/pkg/config"
	"github.com/patrickmn/go-cache"
)

// Record - standard record (struct) for config package
type Record struct {
	Config     *config.Record
	WhoisCache *cache.Cache
}

// CacheControl - cache control commands
type CacheControl struct {
	FlushCache bool
	ListCache  bool
}

// simplest logger, which initialized during starts of the application
var (
	stdlog = log.New(os.Stdout, "[CACHE]: ", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "[CACHE:ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// New - returns new config record initialized with default values
func New(conf *config.Record) *Record {
	WhoisCache := cache.New(
		time.Duration(conf.CacheExpiration)*time.Minute,
		time.Duration(conf.CacheCleanupInterval)*time.Minute,
	)

	return &Record{conf, WhoisCache}
}

// Set - save item in cache
func (memcache *Record) Set(key string, value interface{}) {
	memcache.WhoisCache.Set(key, value, cache.DefaultExpiration)
	stdlog.Println("Items cached: ", memcache.WhoisCache.ItemCount())
}

// Get - get item from cache
func (memcache *Record) Get(key string) (interface{}, bool) {
	_ = memcache.checkCacheControl()
	return memcache.WhoisCache.Get(key)
}

// checkCacheControl - checks and proccess cache control commands
func (memcache *Record) checkCacheControl() error {
	// retrieve cache control execution file info
	_, err := os.Stat(memcache.Config.CacheExecutionFile)
	if err != nil {
		return nil
	}

	// remove cache control execution file
	if err := os.Remove(memcache.Config.CacheExecutionFile); err != nil {
		errlog.Println("Failed to remove cache.control file.", err)
	}

	// get cache control configuration
	control := getCacheControl(memcache.Config)

	// list cache
	if control.ListCache == true {
		memcache.listCache()
	}

	// flush cache
	if control.FlushCache == true {
		stdlog.Printf("Flush cache. %d items removed from cache.", memcache.WhoisCache.ItemCount())
		memcache.WhoisCache.Flush()
	}

	return nil
}

// listCache - create cache list file with all cached items
func (memcache *Record) listCache() {
	items := memcache.WhoisCache.Items()
	if len(items) == 0 {
		return
	}

	listFile, err := os.OpenFile(
		memcache.Config.CacheListFile,
		os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		errlog.Println("Failed to create cache list file.", err)
		return
	}

	if err := listFile.Truncate(0); err != nil {
		errlog.Println("Failed to truncate cache list file. Chache list file is invalid.", err)
		return
	}

	for k := range items {
		_, err := fmt.Fprintln(listFile, k)
		if err != nil {
			errlog.Println("Failed to write into cache list file.", err)
			return
		}
	}

	if err := listFile.Close(); err != nil {
		errlog.Println("Failed to close cache list file.", err)
	}

	stdlog.Printf("Cache list file created. %d items listed.\n", len(items))
}

// initCacheControl - init cache control configuration
func getCacheControl(conf *config.Record) *CacheControl {
	// initialize cache control configuration
	Control := &CacheControl{
		FlushCache: false,
		ListCache:  false,
	}

	// load configuration from file
	if err := loadCacheControlFile(conf, Control); err != nil {
		stdlog.Println("Failed to load cache control configuration file.", err)
	}

	return Control
}

// loadCacheControlFile - loads cache control file into CacheControl record
func loadCacheControlFile(conf *config.Record, control *CacheControl) error {
	// check if cache control configuration file exists
	_, err := os.Stat(conf.CacheControlPath)
	if os.IsNotExist(err) {
		return nil
	}

	// open cache control configuration file
	file, err := os.Open(conf.CacheControlPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// map JSON to struct
	if err := json.NewDecoder(bufio.NewReader(file)).Decode(&control); err != nil {
		return err
	}

	return nil
}

// WriteCacheControl - writes cache control configuration
func WriteCacheControl(conf *config.Record, flush bool, list bool) error {
	conf.Load()

	if conf.CacheEnabled != true {
		return errors.New("ERROR: Cache is disabled")
	}

	control := &CacheControl{
		FlushCache: flush,
		ListCache:  list,
	}

	// convert cache control struct to JSON
	file, err := json.MarshalIndent(control, "", " ")
	if err != nil {
		errlog.Println("Failed to convert cache control configuration to JSON.", err)
		return err
	}

	// save cache control configuration into file
	if err := ioutil.WriteFile(conf.CacheControlPath, file, 0644); err != nil {
		errlog.Println("Failed to create cache control configuration file.", err)
		return err
	}

	// creates cache control execution file
	ctrlFile, err := os.OpenFile(conf.CacheExecutionFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		errlog.Println("Failed to create cache control configuration file", err)
		return err
	}
	defer ctrlFile.Close()

	return nil
}
