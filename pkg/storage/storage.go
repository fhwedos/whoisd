// Copyright 2017 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fhwedos/whoisd/pkg/config"
	"github.com/fhwedos/whoisd/pkg/mapper"
	"github.com/fhwedos/whoisd/pkg/memcache"
	"golang.org/x/net/idna"
)

// Storage - the interface for every implementation of storage
type Storage interface {
	Search(name []string, query []string) (map[string][]string, error)
	SearchRelated(typeTable string, name []string, query []string) (map[string][]string, error)
	SearchMultiple(typeTable string, name []string, query []string) (map[string][]string, error)
}

// Record - standard record (struct) for storage package
type Record struct {
	dataStore Storage
	mapper.Bundle
	conf     *config.Record
	memcache *memcache.Record
}

// simplest logger, which initialized during starts of the application
var (
	stdlog = log.New(os.Stdout, "[STORAGE]: ", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "[STORAGE:ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// New - returns new Storage instance
func New(conf *config.Record, bundle mapper.Bundle, memcache *memcache.Record) *Record {
	switch strings.ToLower(conf.Storage.StorageType) {
	case "mysql":
		return &Record{
			&MysqlRecord{
				conf.Storage.Host,
				conf.Storage.Port,
				conf.Storage.User,
				conf.Storage.Password,
				conf.Storage.IndexBase,
				conf.Storage.TypeTable,
			},
			bundle,
			conf,
			memcache,
		}
	case "elasticsearch":
		return &Record{
			&ElasticsearchRecord{
				conf.Storage.Host,
				conf.Storage.Port,
				conf.Storage.IndexBase,
				conf.Storage.TypeTable,
			},
			bundle,
			conf,
			memcache,
		}
	case "dummy":
		fallthrough
	default:
		return &Record{
			&DummyRecord{conf.Storage.TypeTable},
			bundle,
			conf,
			memcache,
		}
	}
}

// Search and sort a data from the storage
func (storage *Record) Search(query string) (answer string, ok bool, fromCache bool) {
	ok = false
	answer = fmt.Sprintf("No match for domain \"%s\".\n", strings.TrimSpace(query))
	stdlog.Println("query:", query)
	if len(strings.TrimSpace(query)) == 0 {
		errlog.Println("Empty query")
	} else {
		// try to load from cache
		if storage.conf.CacheEnabled == true {
			centry, found := storage.memcache.Get(strings.TrimSpace(query))
			if found {
				cacheResult := centry.(string)
				if answer == cacheResult {
					ok = false
				} else {
					answer = cacheResult
					ok = true
				}

				return answer, ok, true
			}
		}

		entry, err := storage.request(strings.TrimSpace(query))
		if err != nil {
			errlog.Println("Query:", query, err.Error())
		} else {
			if entry == nil || len(entry.Fields) == 0 {
				// save answer in cache
				if storage.conf.CacheEnabled == true {
					//storage.memcache.Set(strings.TrimSpace(query), answer)
					storage.memcache.Set(strings.TrimSpace(query), answer)
				}

				return answer, ok, false
			}
			ok = true

			// get keys of a map and sort their
			keys := make([]string, 0, len(entry.Fields))
			for key := range entry.Fields {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			answer = prepareAnswer(entry, keys)
		}
	}

	// save answer in cache
	if storage.conf.CacheEnabled == true {
		storage.memcache.Set(strings.TrimSpace(query), answer)
	}

	return answer, ok, false
}

// request - get and load bundle by query
func (storage *Record) request(query string) (*mapper.Entry, error) {
	TLD := detachTLD(query)
	if TLD == "" {
		return nil, nil
	}
	template := storage.Bundle.EntryByTLD(TLD)
	if template == nil {
		return nil, nil
	}

	var err error

	entry := new(mapper.Entry)
	entry.Fields = make(map[string]mapper.Field)
	baseField := make(map[string][]string)
	relatedField := make(map[string]map[string][]string)

	// Loads fields with constant values
	for index, field := range template.Fields {
		if len(field.Value) != 0 && len(field.Related) == 0 &&
			len(field.RelatedBy) == 0 && len(field.RelatedTo) == 0 {
			entry.Fields[index] = mapper.Field{
				Key:        field.Key,
				Value:      field.Value,
				Format:     field.Format,
				Multiple:   field.Multiple,
				Hide:       field.Hide,
				ReplacedBy: field.ReplacedBy,
			}
		}
	}

	// Loads base fields (non related)
	for index, field := range template.Fields {
		// Detect base field
		if len(field.Value) == 0 && len(field.Related) != 0 &&
			(len(field.RelatedBy) == 0 || len(field.RelatedTo) == 0) {
			// if baseField not loaded, do it
			if len(baseField) == 0 {
				baseField, err = storage.dataStore.Search(field.Related, []string{query})
				if err != nil {
					return nil, err
				}
				if len(baseField) == 0 {
					return nil, nil
				}
			}
			value := []string{}

			// collects all values into value
			for _, item := range field.Name {
				if result, ok := baseField[item]; ok {
					value = append(value, result...)
				}
			}

			entry.Fields[index] = mapper.Field{
				Key:        field.Key,
				Value:      value,
				Format:     field.Format,
				Multiple:   field.Multiple,
				Hide:       field.Hide,
				ReplacedBy: field.ReplacedBy,
			}
		}
	}

	// Loads related records
	for index, field := range template.Fields {
		// Detect related fields
		if len(field.RelatedBy) != 0 && len(field.RelatedTo) != 0 && len(field.Related) != 0 {
			value := []string{}
			queryRelated := []string{}

			for i := 0; i < len(field.Related); i++ {
				queryRelated = append(queryRelated, baseField[field.Related[i]]...)
			}

			// if non-related field from specified type/table
			if len(field.Value) != 0 {
				queryRelated = []string{field.Value[0]}
			}

			fieldRelatedKey := strings.Join(field.Related, "")

			// if field not cached, do it
			if _, ok := relatedField[fieldRelatedKey]; ok == false {
				if field.Multiple {
					relatedField[fieldRelatedKey], err = storage.dataStore.SearchMultiple(
						field.RelatedTo,
						field.RelatedBy,
						queryRelated,
					)
					if err != nil {
						return nil, err
					}
				} else {
					relatedField[fieldRelatedKey], err = storage.dataStore.SearchRelated(
						field.RelatedTo,
						field.RelatedBy,
						queryRelated,
					)
					if err != nil {
						return nil, err
					}
				}
			}
			// collects all values into value
			for _, item := range field.Name {
				if result, ok := relatedField[fieldRelatedKey][item]; ok {
					value = append(value, result...)
				}
			}
			entry.Fields[index] = mapper.Field{
				Key:        field.Key,
				Value:      value,
				Format:     field.Format,
				Multiple:   field.Multiple,
				Hide:       field.Hide,
				ReplacedBy: field.ReplacedBy,
			}
		}
	}

	return entry, nil
}

// detachTLD - isolates TLD part from a query
func detachTLD(query string) (TLD string) {
	parts := strings.Split(query, ".")
	if len(parts) > 1 {
		TLD = parts[len(parts)-1]
	}
	return
}

// prepares join and multiple actions in the answer
func prepareAnswer(entry *mapper.Entry, keys []string) (answer string) {
	for _, index := range keys {
		key := entry.Fields[index].Key
		if strings.Contains(entry.Fields[index].Format, "{idn}") {
			entry.Fields[index] = decodeIDN(entry.Fields[index])
		}
		if entry.Fields[index].Multiple {
			for _, value := range entry.Fields[index].Value {
				if entry.Fields[index].Hide && value == "" ||
					entry.Fields[index].ReplacedBy != "" && isNotEmptyField(entry, entry.Fields[index].ReplacedBy) {
					continue
				}
				if entry.Fields[index].Format != "" {
					value = handleTags(entry.Fields[index].Format, []string{value})
				}
				answer = strings.Join([]string{answer, key, value, "\n"}, "")
			}
		} else {
			var value string
			if entry.Fields[index].Format != "" {
				value = handleTags(entry.Fields[index].Format, entry.Fields[index].Value)
			} else {
				value = strings.Trim(strings.Join(entry.Fields[index].Value, " "), " ")
			}
			if entry.Fields[index].Hide && value == "" ||
				entry.Fields[index].ReplacedBy != "" && isNotEmptyField(entry, entry.Fields[index].ReplacedBy) {
				continue
			}
			answer = strings.Join([]string{answer, key, value, "\n"}, "")
		}
	}

	return answer
}

func isNotEmptyField(entry *mapper.Entry, index string) bool {
	for _, value := range entry.Fields[index].Value {
		if value != "" {
			return true
		}
	}
	return false
}

// decodes IDN names to Unicode and adds it to value
func decodeIDN(field mapper.Field) mapper.Field {
	for _, item := range field.Value {
		idnItem, err := idna.ToUnicode(item)
		if err == nil && idnItem != item {
			field.Value = append(
				field.Value,
				strings.Replace(field.Format, "{idn}", idnItem, 1),
			)
		}
	}
	field.Format = ""
	return field
}

// handle all tags defined in format string
func handleTags(format string, value []string) string {
	// template of date to parse
	const (
		shortDateFormat = "2006.01.02"
		longDateFormat  = "2006-01-02 15:04:05"
	)
	for _, item := range value {
		if strings.Contains(format, "{date}") || strings.Contains(format, "{shortdate}") {
			buildTime, err := time.Parse(longDateFormat, item)
			if err != nil && len(strings.TrimSpace(item)) == 0 {
				buildTime = time.Now()
			}
			if strings.Contains(format, "{date}") {
				format = strings.Replace(format, "{date}", buildTime.Format(time.RFC3339), 1)
			} else {
				format = strings.Replace(format, "{shortdate}", buildTime.Format(shortDateFormat), 1)
			}
		}
		format = strings.Replace(format, "{string}", item, 1)
	}
	format = strings.NewReplacer("{string}", "", "{date}", "", "{shortdate}", "").Replace(format)

	return strings.Trim(format, ". ")
}
