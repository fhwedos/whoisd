package storage

import (
	"testing"
)

func TestMysqlSearchEmpty(t *testing.T) {
	dummy := MysqlRecord{"localhost", 3760, "whois", "domain", "test", "test"}
	var emptyResult map[string][]string
	var err error
	emptyResult, err = dummy.Search([]string{"name"}, []string{""})
	if err == nil {
		t.Error("Expected error for empty query, got", err)
	}
	emptyResult, err = dummy.Search([]string{"name"}, []string{"aaa"})
	if len(emptyResult) != 0 {
		t.Error("Expected len of empty query", 0, ", got", len(emptyResult))
	}
	emptyResult, err = dummy.SearchRelated("customer", []string{""}, []string{""})
	if err == nil {
		t.Error("Expected error for empty query, got", err)
	}
	emptyResult, err = dummy.SearchRelated("customer", []string{"handle"}, []string{"AA-BB"})
	if len(emptyResult) != 0 {
		t.Error("Expected len of empty query", 0, ", got", len(emptyResult))
	}
	emptyResult, err = dummy.SearchMultiple("nameserver", []string{""}, []string{""})
	if err == nil {
		t.Error("Expected error for empty query, got", err)
	}
	emptyResult, err = dummy.SearchMultiple("nameserver", []string{"nsgroupId"}, []string{"7"})
	if len(emptyResult) != 0 {
		t.Error("Expected len of empty query", 0, ", got", len(emptyResult))
	}
}
