package dataset

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/qri-io/cafs"
	"github.com/qri-io/cafs/memfile"
)

// Query defines an action to be taken on one or more structures
type Query struct {
	// private storage for reference to this object
	path datastore.Key

	// Syntax is an identifier string for the statement syntax (Eg, "SQL")
	Syntax string `json:"syntax"`
	// Structures is a map of all structures referenced in this query,
	// with alphabetical keys generated by datasets in order of appearance within the query.
	// Keys are _always_ referenced in the form [a-z,aa-zz,aaa-zzz, ...] by order of appearence.
	// The query itself is rewritten to refer to these table names using bind variables
	Structures map[string]*Structure `json:"structures"`
	// Statement is the is parsed & rewritten to a _standard form_ to maximize hash overlap.
	// Writing a query to it's standard form involves making deterministic choices to
	// remove non-semantic whitespace, rewrite semantically-equivalent terms like "&&" and "AND"
	// to a chosen version, et cetera.
	// Greater precision of querying format will increase the chances of hash discovery.
	Statement string `json:"statement"`
	// Structure is a path to an algebraic structure that is the _output_ of this structure
	Structure *Structure
}

// _query is a private struct for marshaling into & out of.
// fields must remain sorted in lexographical order
type _query struct {
	Structure  *Structure            `json:"outputStructure"`
	Statement  string                `json:"statement"`
	Structures map[string]*Structure `json:"structures"`
	Syntax     string                `json:"syntax"`
}

// MarshalJSON satisfies the json.Marshaler interface
func (q Query) MarshalJSON() ([]byte, error) {
	// if we're dealing with an empty object that has a path specified, marshal to a string instead
	if q.path.String() != "" && q.Structure == nil && q.Syntax == "" && q.Structures == nil {
		return q.path.MarshalJSON()
	}

	return json.Marshal(&_query{
		Structure:  q.Structure,
		Statement:  q.Statement,
		Structures: q.Structures,
		Syntax:     q.Syntax,
	})
}

// UnmarshalJSON satisfies the json.Unmarshaler interface
func (q *Query) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*q = Query{path: datastore.NewKey(s)}
		return nil
	}

	_q := &_query{}
	if err := json.Unmarshal(data, _q); err != nil {
		return err
	}

	*q = Query{
		Structures: _q.Structures,
		Structure:  _q.Structure,
		Statement:  _q.Statement,
		Syntax:     _q.Syntax,
	}
	return nil
}

func (q *Query) IsEmpty() bool {
	return q.Statement == "" && q.Syntax == "" && q.Structure == nil && q.Structures == nil
}

// func (q *Query) LoadStructures(store datastore.Datastore) (structs map[string]*Structure, err error) {
// 	structs = map[string]*Structure{}
// 	for key, path := range q.Structures {
// 		s, err := LoadStructure(store, path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		structs[key] = s
// 	}
// 	return
// }

// func (q *Query) LoadAbstractStructures(store datastore.Datastore) (structs map[string]*Structure, err error) {
// 	structs = map[string]*Structure{}
// 	for key, path := range q.Structures {
// 		s, err := LoadStructure(store, path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		structs[key] = s.Abstract()
// 	}
// 	return
// }

// LoadQuery loads a query from a given path in a store
func LoadQuery(store cafs.Filestore, path datastore.Key) (q *Query, err error) {
	q = &Query{path: path}
	err = q.Load(store)
	return
}

// UnmarshalResource tries to extract a resource type from an empty
// interface. Pairs nicely with datastore.Get() from github.com/ipfs/go-datastore
func UnmarshalQuery(v interface{}) (*Query, error) {
	switch q := v.(type) {
	case *Query:
		return q, nil
	case Query:
		return &q, nil
	case []byte:
		query := &Query{}
		err := json.Unmarshal(q, query)
		return query, err
	default:
		return nil, fmt.Errorf("couldn't parse query")
	}
}

func (q *Query) Load(store cafs.Filestore) error {
	if q.path.String() == "" {
		return ErrNoPath
	}
	v, err := store.Get(q.path)
	if err != nil {
		return err
	}

	uq, err := UnmarshalQuery(v)
	if err != nil {
		return err
	}

	*q = *uq
	return nil
}

func (q *Query) Save(store cafs.Filestore, pin bool) (datastore.Key, error) {
	if q == nil {
		return datastore.NewKey(""), nil
	}

	// *don't* need to break query out into different structs.
	// stpath, err := q.Structure.Save(store)
	// if err != nil {
	// 	return datastore.NewKey(""), err
	// }

	qdata, err := json.Marshal(q)
	if err != nil {
		return datastore.NewKey(""), err
	}

	return store.Put(memfile.NewMemfileBytes("query.json", qdata), pin)
}
