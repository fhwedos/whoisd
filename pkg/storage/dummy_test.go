package storage

import (
	"testing"
)

func TestDummySearch(t *testing.T) {

	type testData struct {
		query    []string
		name     []string
		ask      string
		expected []string
	}

	var tests = []testData{
		{[]string{"google.com"}, []string{"name"}, "ownerHandle", []string{"MMR-2383"}},
		{[]string{"google.com"}, []string{"name"}, "updatedDate", []string{"2014-05-19 04:00:17"}},
		{[]string{"google.com"}, []string{"name"}, "dnssec", []string{"unsigned"}},
		{[]string{"example.tld"}, []string{"name"}, "techHandle", []string{"5372811-ERL"}},
		{[]string{"example.tld"}, []string{"name"}, "domainStatus",
			[]string{
				"clientDeleteProhibited",
				"clientRenewProhibited",
				"clientTransferProhibited",
			},
		},
		{[]string{"example.tld"}, []string{"name"}, "dnssec", []string{"signedDelegation"}},
	}

	dummy := DummyRecord{"domain"}
	for _, data := range tests {
		result, err := dummy.Search(data.name, data.query)
		if err != nil {
			t.Error(err.Error())
		} else {
			if len(result) == 0 {
				t.Error("Empty set for", data.query)
			}
			for index, item := range result[data.ask] {
				if item != data.expected[index] {
					t.Error("Expected", data.expected, ", got", item)
				}
			}
		}
	}
}

func TestDummySearchRelated(t *testing.T) {

	type testData struct {
		query     []string
		name      []string
		relatedTo string
		ask       string
		expected  []string
	}

	var tests = []testData{
		{[]string{"MMR-2383"}, []string{"handle"}, "customer", "address.street", []string{"1600 Amphitheatre Parkway"}},
		{[]string{"MMR-2383"}, []string{"handle"}, "customer", "email", []string{"dns-admin@google.com"}},
		{[]string{"MMR-2383"}, []string{"handle"}, "customer", "name.lastName", []string{"Admin"}},
		{[]string{"MMA-2211"}, []string{"handle"}, "customer", "address.street", []string{"2400 E. Bayshore Pkwy"}},
	}

	dummy := DummyRecord{"domain"}
	for _, data := range tests {
		result, err := dummy.SearchRelated(data.relatedTo, data.name, data.query)
		if err != nil {
			t.Error(err.Error())
		} else {
			if len(result) == 0 {
				t.Error("Empty set for", data.query)
			}
			for index, item := range result[data.ask] {
				if item != data.expected[index] {
					t.Error("Expected", data.expected[index], ", got", item)
				}
			}
		}
	}
}

func TestDummySearchMultiple(t *testing.T) {

	type testData struct {
		query     []string
		name      []string
		relatedTo string
		ask       string
		expected  []string
	}

	var tests = []testData{
		{[]string{"1"}, []string{"nsgroupId"}, "nameserver", "name",
			[]string{
				"NS01.EXAMPLE-REGISTRAR.TLD",
				"NS02.EXAMPLE-REGISTRAR.TLD",
			},
		},
		{[]string{"2"}, []string{"nsgroupId"}, "nameserver", "name",
			[]string{
				"ns1.google.com",
				"ns2.google.com",
				"ns3.google.com",
				"ns4.google.com",
			},
		},
	}

	dummy := DummyRecord{"domain"}
	for _, data := range tests {
		result, err := dummy.SearchMultiple(data.relatedTo, data.name, data.query)
		if err != nil {
			t.Error(err.Error())
		} else {
			if len(result) == 0 {
				t.Error("Empty set for", data.query)
			}
			if len(result[data.ask]) != len(data.expected) {
				t.Error("No multiple records, expected", len(data.expected), ", got", len(result[data.ask]))
			}
			for index, item := range result[data.ask] {
				if item != data.expected[index] {
					t.Error("Expected", data.expected[index], ", got", item)
				}
			}
		}
	}
}

func TestDummySearchEmpty(t *testing.T) {
	dummy := DummyRecord{"domain"}
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
