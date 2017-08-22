package dataset

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/qri-io/castore"
	"time"
)

// Dataset is stored separately from prescriptive metadata stored in Resource structs
// to maximize overlap of the formal query & resource definitions.
// A Dataset must resolve to one and only one entity, specified by a `data` property.
// It's structure must be specified by a structure definition.
// This also creates space for subjective claims about datasets, and allows metadata
// to take on a higher frequency of change in contrast to the underlying definition.
// In addition, descriptive metadata can and should be author attributed
// associating descriptive claims about a resource with a cyptographic keypair which
// may represent a person, group of people, or software.
// This metadata format is also subject to massive amounts of change.
// Design goals should include making this compatible with the DCAT spec,
// with the one major exception that hashes are acceptable in place of urls.
type Dataset struct {
	// private storage for reference to this object
	path datastore.Key

	// Time this dataset was created. Required. Datasets are immutable, so no "updated"
	Timestamp time.Time `json:"timestamp"`
	// Structure of this dataset, required
	Structure *Structure `json:"structure"`

	// Data is the path to the hash of raw data as it resolves on the network.
	Data datastore.Key `json:"data"`
	// Length is the length of the data object in bytes.
	// must always match & be present
	Length int `json:"length"`
	// Previous connects datasets to form a historical DAG
	Previous *Dataset `json:"previous,omitempty"`

	// Title of this dataset
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
	// path to readme
	Readme datastore.Key `json:"readme,omitempty"`
	// Author
	Author      *User       `json:"author,omitempty"`
	Citations   []*Citation `json:"citations"`
	Image       string      `json:"image,omitempty"`
	Description string      `json:"description,omitempty"`
	Homepage    string      `json:"homepage,omitempty"`
	IconImage   string      `json:"icon_image,omitempty"`
	//
	PosterImage string `json:"poster_image,omitempty"`
	// License
	License *License `json:"license,omitempty"`
	// SemVersion this dataset?
	Version VersionNumber `json:"version,omitempty"`
	// String of Keywords
	Keywords []string `json:"keywords,omitempty"`
	// Contribute
	Contributors []*User `json:"contributors,omitempty"`
	// QueryString is the user-inputted string of this query
	QueryString string `json:"queryString,omitempty"`
	// Query is a path to a query that generated this resource
	Query *Query `json:"query,omitempty"`
	// Syntax this query was written in
	QuerySyntax string `json:"querySyntax"`
	// queryPlatform is an identifier for the operating system that performed the query
	QueryPlatform string `json:"queryPlatform,omitempty"`
	// QueryEngine is an identifier for the application that produced the result
	QueryEngine string `json:"queryEngine,omitempty"`
	// QueryEngineConfig outlines any configuration that would affect the resulting hash
	QueryEngineConfig map[string]interface{} `json:"queryEngineConfig,omitempty`
	// Resources is a map of dataset names to dataset references this query is derived from
	// all tables referred to in the query should be present here
	Resources map[string]*Dataset `json:"resources,omitempty"`
	// meta holds additional arbitrarty metadata not covered by the spec
	// when encoding & decoding json values here will be hoisted into the
	// Dataset object
	meta map[string]interface{}
}

// Meta gives access to additional metadata not covered by dataset metadata
func (d *Dataset) Meta() map[string]interface{} {
	if d.meta == nil {
		d.meta = map[string]interface{}{}
	}
	return d.meta
}

// func (d *Dataset) LoadStructure(store datastore.Datastore) (*Structure, error) {
// 	if d.Structure != nil && d.Structure.path != "" {
// 		return LoadStructure(store, d.Structure.path)
// 	}
// 	return
// }

func (d *Dataset) LoadData(store datastore.Datastore) ([]byte, error) {
	v, err := store.Get(d.Data)
	if err != nil {
		return nil, err
	}

	if data, ok := v.([]byte); ok {
		return data, nil
	}

	return nil, fmt.Errorf("wrong data type for dataset data: %s", d.Data)
}

// MarshalJSON uses a map to combine meta & standard fields.
// Marshalling a map[string]interface{} automatically alpha-sorts the keys.
func (d *Dataset) MarshalJSON() ([]byte, error) {
	// if we're dealing with an empty object that has a path specified, marshal to a string instead
	// TODO - check all fields
	if d.path.String() != "" && d.Title == "" && d.Description == "" && d.Structure == nil && d.Timestamp.IsZero() && d.Previous == nil {
		return d.path.MarshalJSON()
	}

	data := d.Meta()

	// required fields first
	data["title"] = d.Title
	data["timestamp"] = d.Timestamp
	data["data"] = d.Data
	data["length"] = d.Length
	data["structure"] = d.Structure

	if d.Previous != nil {
		data["previous"] = d.Previous
	}
	if d.Url != "" {
		data["url"] = d.Url
	}
	if d.Readme.String() != "" {
		data["readme"] = d.Readme
	}
	if d.Author != nil {
		data["author"] = d.Author
	}
	if d.Image != "" {
		data["image"] = d.Image
	}
	if d.Description != "" {
		data["description"] = d.Description
	}
	if d.Homepage != "" {
		data["homepage"] = d.Homepage
	}
	if d.IconImage != "" {
		data["iconImage"] = d.IconImage
	}
	if d.PosterImage != "" {
		data["posterImage"] = d.PosterImage
	}
	if d.License != nil {
		data["license"] = d.License
	}
	if d.Version != VersionNumber("") {
		data["version"] = d.Version
	}
	if d.Keywords != nil {
		data["keywords"] = d.Keywords
	}
	if d.Contributors != nil {
		data["contributors"] = d.Contributors
	}
	if d.Citations != nil {
		data["citations"] = d.Citations
	}

	if d.QueryString != "" {
		data["queryString"] = d.QueryString
	}
	if d.Query != nil {
		data["query"] = d.Query
	}
	if d.QueryPlatform != "" {
		data["querySyntax"] = d.QuerySyntax
	}
	if d.QueryPlatform != "" {
		data["queryPlatform"] = d.QueryPlatform
	}
	if d.QueryEngine != "" {
		data["queryEngine"] = d.QueryEngine
	}
	if d.QueryEngineConfig != nil {
		data["queryEngineConfig"] = d.QueryEngineConfig
	}
	if d.Resources != nil {
		data["resources"] = d.Resources
	}

	return json.Marshal(data)
}

// internal struct for json unmarshaling
type _dataset Dataset

// UnmarshalJSON implements json.Unmarshaller
func (d *Dataset) UnmarshalJSON(data []byte) error {
	// first check to see if this is a valid path ref
	var path string
	if err := json.Unmarshal(data, &path); err == nil {
		*d = Dataset{path: datastore.NewKey(path)}
		return nil
	}

	// TODO - I'm guessing what follows could be better
	ds := _dataset{}
	if err := json.Unmarshal(data, &ds); err != nil {
		return err
	}

	meta := map[string]interface{}{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}

	for _, f := range []string{
		"title",
		"url",
		"readme",
		"author",
		"image",
		"structure",
		"citations",
		"description",
		"homepage",
		"iconImage",
		"posterImage",
		"license",
		"version",
		"keywords",
		"contributors",
		"meta",
		"timestamp",
		"length",
		"previous",
		"data",
		"query",
		"abstract",
		"queryPlatform",
		"queryEngine",
		"queryEngineConfig",
		"resources",
	} {
		delete(meta, f)
	}

	ds.meta = meta
	*d = Dataset(ds)
	return nil
}

// LoadDataset loads a dataset from a given path in a store
func LoadDataset(store datastore.Datastore, path datastore.Key) (*Dataset, error) {
	v, err := store.Get(path)
	if err != nil {
		return nil, err
	}

	return UnmarshalDataset(v)
}

// UnmarshalDataset tries to extract a dataset type from an empty
// interface. Pairs nicely with datastore.Get() from github.com/ipfs/go-datastore
func UnmarshalDataset(v interface{}) (*Dataset, error) {
	switch r := v.(type) {
	case *Dataset:
		return r, nil
	case Dataset:
		return &r, nil
	case []byte:
		dataset := &Dataset{}
		err := json.Unmarshal(r, dataset)
		return dataset, err
	default:
		return nil, fmt.Errorf("couldn't parse dataset")
	}
}

// // ReferencedDataset is a dataset with all references as datastore.Key's
// type ReferencedDataset struct {
// 	Timestamp         time.Time                `json:"timestamp"`
// 	Structure         datastore.Key            `json:"structure"`
// 	Data              datastore.Key            `json:"data"`
// 	Length            int                      `json:"length"`
// 	Previous          datastore.Key            `json:"previous,omitempty"`
// 	Title             string                   `json:"title,omitempty"`
// 	Url               string                   `json:"url,omitempty"`
// 	Readme            datastore.Key            `json:"readme,omitempty"`
// 	Author            *User                    `json:"author,omitempty"`
// 	Citations         []*Citation              `json:"citations"`
// 	Image             string                   `json:"image,omitempty"`
// 	Description       string                   `json:"description,omitempty"`
// 	Homepage          string                   `json:"homepage,omitempty"`
// 	IconImage         string                   `json:"icon_image,omitempty"`
// 	PosterImage       string                   `json:"poster_image,omitempty"`
// 	License           *License                 `json:"license,omitempty"`
// 	Version           VersionNumber            `json:"version,omitempty"`
// 	Keywords          []string                 `json:"keywords,omitempty"`
// 	Contributors      []*User                  `json:"contributors,omitempty"`
// 	QueryString       string                   `json:"queryString,omitempty"`
// 	Query             datastore.Key            `json:"query,omitempty"`
// 	QuerySyntax       string                   `json:"querySyntax"`
// 	QueryPlatform     string                   `json:"queryPlatform,omitempty"`
// 	QueryEngine       string                   `json:"queryEngine,omitempty"`
// 	QueryEngineConfig map[string]interface{}   `json:"queryEngineConfig,omitempty`
// 	Resources         map[string]datastore.Key `json:"resources,omitempty"`
// 	meta              map[string]interface{}
// }

func (ds *Dataset) Load(store castore.Datastore, path datastore.Key) error {
	v, err := store.Get(path)
	if err != nil {
		return err
	}

	d, err := UnmarshalDataset(v)
	if err != nil {
		return err
	}

	*ds = *d
	return nil
}

func (ds *Dataset) Save(store castore.Datastore) (datastore.Key, error) {
	// if ds == nil {
	// 	return datastore.NewKey(""), nil
	// }

	// Ah fuck it, for now let's just write the whole thing to JSON
	dsdata, err := json.Marshal(ds)
	if err != nil {
		return datastore.NewKey(""), err
	}

	return store.Put(dsdata)
}
