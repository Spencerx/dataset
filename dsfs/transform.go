package dsfs

import (
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/qri-io/cafs"
	"github.com/qri-io/dataset"
)

// LoadTransform loads a transform from a given path in a store
func LoadTransform(store cafs.Filestore, path datastore.Key) (q *dataset.Transform, err error) {
	path = PackageKeypath(store, path, PackageFileTransform)
	return loadTransform(store, path)
}

// loadTransform assumes the provided path is correct
func loadTransform(store cafs.Filestore, path datastore.Key) (q *dataset.Transform, err error) {
	data, err := fileBytes(store.Get(path))
	if err != nil {
		log.Debug(err.Error())
		return nil, fmt.Errorf("error loading transform raw data: %s", err.Error())
	}

	return dataset.UnmarshalTransform(data)
}

// SaveTransform writes a transform to a cafs
func SaveTransform(store cafs.Filestore, q *dataset.Transform, pin bool) (path datastore.Key, err error) {
	// copy transform
	save := &dataset.Transform{}
	save.Assign(q)
	save.Qri = dataset.KindTransform

	if q.Structure != nil && !q.Structure.IsEmpty() {
		path, err := SaveStructure(store, q.Structure, pin)
		if err != nil {
			log.Debug(err.Error())
			return datastore.NewKey(""), err
		}
		save.Structure = dataset.NewStructureRef(path)
	}

	tf, err := JSONFile(PackageFileTransform.String(), save)
	if err != nil {
		log.Debug(err.Error())
		return datastore.NewKey(""), fmt.Errorf("error marshaling transform data to json: %s", err.Error())
	}

	return store.Put(tf, pin)
}

// ErrNoTransform is the error for asking a dataset without a tranform component for viz info
var ErrNoTransform = fmt.Errorf("this dataset has no transform component")

// LoadTransformScript loads transform script data from a dataset path if the given dataset has a transform script specified
// the returned cafs.File will be the value of dataset.Transform.ScriptPath
func LoadTransformScript(store cafs.Filestore, dspath datastore.Key) (cafs.File, error) {
	ds, err := LoadDataset(store, dspath)
	if err != nil {
		return nil, err
	}

	if ds.Transform == nil || ds.Transform.ScriptPath == "" {
		return nil, ErrNoTransform
	}

	return store.Get(datastore.NewKey(ds.Transform.ScriptPath))
}
