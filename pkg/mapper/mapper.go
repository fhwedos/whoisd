// Copyright 2017 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package mapper

// Bundle - set of entries
type Bundle []Entry

// Entry contains whois data specified for TLDs list
type Entry struct {
	// TLDs - list of TLDs, which accepted by specified Entry
	TLDs []string

	// Default - use this entry as default for all undefined TLDs
	Default bool

	// a list of fields from "01" to last number "nn" in ascending order
	Fields map[string]Field
}

// Field - representation of one field
type Field struct {
	// a label in whois output
	Key string

	// used if a field has constant value
	Value []string

	// a name of the field in the database, if "value" is not defined
	Name []string

	// special instructions to indicate how to display the field
	Format string

	// if this option is set to 'true', each value will be repeated in whois output with the same label
	Multiple bool

	// if this option is set to 'true', a value of the field will not shown in whois output
	Hide bool

	// a name of the field in a database through which a request for
	Related []string

	// a name of the field in a database through which related a request for
	RelatedBy []string

	// a name of the table/type in a database through which made a relation
	RelatedTo string

	// it contains the number of field that replaces this field.
	// If the field is referenced by "replacedBy" has non-empty value,
	// then this field will not show in whois output, because this field
	// should be replaced by the field specified in 'ReplacedBy'
	ReplacedBy string
}

// EntryByTLD returns whois data by TLD
func (bundle Bundle) EntryByTLD(TLD string) *Entry {
	var defaultIndex int
	for index := range bundle {
		for _, item := range bundle[index].TLDs {
			if item == TLD {
				return &bundle[index]
			}
			if bundle[index].Default {
				defaultIndex = index
			}
		}
	}
	if defaultIndex != 0 {
		return &bundle[defaultIndex]
	}
	return nil
}
