package dataset

import (
	"encoding/json"
	"github.com/ipfs/go-datastore"
	// "gx/ipfs/QmVSase1JP7cq9QkPT46oNwdp9pT6kBkG3oqS14y3QcZjG/go-datastore"
)

// Query defines an action to be taken on one or more resources
type Query struct {
	// Syntax is an identifier string for the statement syntax (Eg, "SQL")
	Syntax string `json:"syntax"`
	// Resources is a map of all datasets referenced in this query,
	// with alphabetical keys generated by datasets in order of appearance within the query.
	// Keys are _always_ referenced in the form [a-z,aa-zz,aaa-zzz, ...] by order of appearence.
	// The query itself is rewritten to refer to these table names using bind variables
	Resources map[string]datastore.Key `json:"resources"`
	// Statement is the is parsed & rewritten to a _standard form_ to maximize hash overlap.
	// Writing a query to it's standard form involves making deterministic choices to
	// remove non-semantic whitespace, rewrite semantically-equivalent terms like "&&" and "AND"
	// to a chosen version, et cetera.
	// Greater precision of querying format will increase the chances of hash discovery.
	Statement string `json:"statement"`
	// Schema defines the intended output schema of the query
	Schema *Schema `json:"schema"`
}

// _query is a private struct for marshaling into & out of.
// fields must remain sorted in lexographical order
type _query struct {
	Resources map[string]datastore.Key `json:"resources"`
	Schema    *Schema                  `json:"schema"`
	Statement string                   `json:"statement"`
	Syntax    string                   `json:"syntax"`
}

// MarshalJSON satisfies the json.Marshaler interface
func (q Query) MarshalJSON() ([]byte, error) {
	return json.Marshal(&_query{
		Resources: q.Resources,
		Schema:    q.Schema,
		Statement: q.Statement,
		Syntax:    q.Syntax,
	})
}

// UnmarshalJSON satisfies the json.Unmarshaler interface
func (q *Query) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// return fmt.Errorf("Data Type should be a string, got %s", data)
		*q = Query{Statement: s}
		return nil
	}

	_q := &_query{}
	if err := json.Unmarshal(data, _q); err != nil {
		return err
	}

	*q = Query{
		Resources: _q.Resources,
		Schema:    _q.Schema,
		Statement: _q.Statement,
		Syntax:    _q.Syntax,
	}
	return nil
}

// type ResourcesFunc func(int, *Resource) error

// func (ds *Resource) Resources(depth int, fn ResourcesFunc) (err error) {
// 	// call once for base dataset
// 	if err = fn(depth, ds); err != nil {
// 		return
// 	}

// 	depth++
// 	for _, d := range ds.Resources {
// 		if err = d.Resources(depth, fn); err != nil {
// 			return
// 		}
// 	}

// 	return
// }
