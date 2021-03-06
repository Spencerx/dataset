package dsfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/qri-io/dataset"
	"github.com/qri-io/dataset/dstest"
	"github.com/qri-io/qfs/cafs"
)

var Viz1 = &dataset.Viz{
	Format:     "foo",
	Qri:        dataset.KindViz.String(),
	ScriptPath: "bar",
}

func TestLoadViz(t *testing.T) {
	store := cafs.NewMapstore()
	a, err := SaveViz(store, Viz1, true)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if _, err := LoadViz(store, a); err != nil {
		t.Errorf(err.Error())
	}
}

func TestLoadVizScript(t *testing.T) {
	store := cafs.NewMapstore()
	privKey, err := crypto.UnmarshalPrivateKey(testPk)
	if err != nil {
		t.Fatalf("error unmarshaling private key: %s", err.Error())
	}

	_, err = LoadVizScript(store, "")
	if err == nil {
		t.Error("expected load empty key to fail")
	}

	tc, err := dstest.NewTestCaseFromDir("testdata/cities_no_commit_title")
	if err != nil {
		t.Fatal(err.Error())
	}
	path, err := CreateDataset(store, tc.Input, nil, privKey, true, false, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err = LoadVizScript(store, path); err != ErrNoViz {
		t.Errorf("expected no viz script error. got: %s", err)
	}

	tc, err = dstest.NewTestCaseFromDir("testdata/all_fields")
	if err != nil {
		t.Fatal(err.Error())
	}
	vsf, _ := tc.VizScriptFile()
	vizPath, err := store.Put(vsf, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	tc.Input.Viz.ScriptPath = vizPath
	path, err = CreateDataset(store, tc.Input, nil, privKey, true, false, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	file, err := LoadVizScript(store, path)
	if err != nil {
		t.Fatalf("expected viz script to load. got: %s", err)
	}

	vsf, _ = tc.VizScriptFile()

	r := &EqualReader{file, vsf}
	if _, err := ioutil.ReadAll(r); err != nil {
		t.Error(err.Error())
	}
}

var ErrStreamsNotEqual = fmt.Errorf("streams are not equal")

// EqualReader confirms two readers are exactly the same, throwing an error
// if they return
type EqualReader struct {
	a, b io.Reader
}

func (r *EqualReader) Read(p []byte) (int, error) {
	pb := make([]byte, len(p))
	readA, err := r.a.Read(p)
	if err != nil {
		return readA, err
	}

	readB, err := r.b.Read(pb)
	if err != nil {
		return readA, err
	}

	if readA != readB {
		return readA, ErrStreamsNotEqual
	}

	for i, b := range p {
		if pb[i] != b {
			return readA, ErrStreamsNotEqual
		}
	}

	return readA, nil
}
