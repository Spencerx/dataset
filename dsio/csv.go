package dsio

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/qri-io/dataset"
	"github.com/qri-io/dataset/dsio/replacecr"
	"github.com/qri-io/dataset/vals"
)

// CSVReader implements the RowReader interface for the CSV data format
type CSVReader struct {
	st         *dataset.Structure
	readHeader bool
	r          *csv.Reader
	types      []string
}

var _ EntryReader = (*CSVReader)(nil)

// NewCSVReader creates a reader from a structure and read source
func NewCSVReader(st *dataset.Structure, r io.Reader) *CSVReader {
	// TODO - handle error
	_, types, _ := terribleHackToGetHeaderRowAndTypes(st)

	csvr := csv.NewReader(replacecr.Reader(r))

	if fopts, err := dataset.ParseFormatConfigMap(dataset.CSVDataFormat, st.FormatConfig); err == nil {
		if opts, ok := fopts.(*dataset.CSVOptions); ok {
			csvr.LazyQuotes = opts.LazyQuotes
			if opts.VariadicFields == true {
				csvr.FieldsPerRecord = -1
			}
			if opts.Separator != rune(0) {
				csvr.Comma = opts.Separator
			}
		}
	}

	return &CSVReader{
		st:    st,
		r:     csvr,
		types: types,
	}
}

// Structure gives this reader's structure
func (r *CSVReader) Structure() *dataset.Structure {
	return r.st
}

// ReadEntry reads one CSV record from the reader
func (r *CSVReader) ReadEntry() (Entry, error) {
	if !r.readHeader {
		if HasHeaderRow(r.st) {
			if _, err := r.r.Read(); err != nil {
				if err.Error() != "EOF" {
					log.Debug(err.Error())
				}
				return Entry{}, err
			}
		}
		r.readHeader = true
	}

	data, err := r.r.Read()
	if err != nil {
		log.Debug(err.Error())
		return Entry{}, err
	}

	value, err := r.decode(data)
	if err != nil {
		log.Debug(err.Error())
		return Entry{}, err
	}

	return Entry{Value: value}, nil
}

// Close finalizes the reader
func (r *CSVReader) Close() error {
	// TODO (b5): we should retain a reference to the underlying reader &
	// check if it's an io.ReadCloser, calling close here if so
	return nil
}

// decode uses specified types from structure's schema to cast csv string values to their
// intended types. If casting fails because the data is invalid, it's left as a string instead
// of causing an error.
func (r *CSVReader) decode(strings []string) ([]interface{}, error) {
	vs := make([]interface{}, len(strings))
	types := r.types
	if len(types) < len(strings) {
		// TODO - fix. for now is types fails to parse we just assume all types
		// are strings
		types = make([]string, len(strings))
		for i := range types {
			types[i] = "string"
		}
	}
	for i, str := range strings {
		vs[i] = str

		switch types[i] {
		case "number":
			if num, err := vals.ParseNumber([]byte(str)); err == nil {
				vs[i] = num
			}
		case "integer":
			if num, err := vals.ParseInteger([]byte(str)); err == nil {
				vs[i] = num
			}
		case "boolean":
			if b, err := vals.ParseBoolean([]byte(str)); err == nil {
				vs[i] = b
			}
		case "object":
			v := map[string]interface{}{}
			if err := json.Unmarshal([]byte(str), &v); err == nil {
				vs[i] = v
			}
		case "array":
			v := []interface{}{}
			if err := json.Unmarshal([]byte(str), &v); err == nil {
				vs[i] = v
			}
		case "null":
			vs[i] = nil
		}
	}

	return vs, nil
}

// HasHeaderRow checks Structure for the presence of the HeaderRow flag
func HasHeaderRow(st *dataset.Structure) bool {
	if st.DataFormat() == dataset.CSVDataFormat && st.FormatConfig != nil {
		if csvOpt, err := dataset.NewCSVOptions(st.FormatConfig); err == nil {
			return csvOpt.HeaderRow
		}
	}
	return false
}

// CSVWriter implements the RowWriter interface for
// CSV-formatted data
type CSVWriter struct {
	rowsWritten int
	w           *csv.Writer
	st          *dataset.Structure
	types       []string
}

// NewCSVWriter creates a Writer from a structure and write destination
func NewCSVWriter(st *dataset.Structure, w io.Writer) *CSVWriter {
	// TODO - capture error
	titles, types, _ := terribleHackToGetHeaderRowAndTypes(st)

	writer := csv.NewWriter(w)
	opts, err := dataset.NewCSVOptions(st.FormatConfig)
	if opts != nil && err == nil {
		if opts.Separator != rune(0) {
			writer.Comma = opts.Separator
		}
	}

	wr := &CSVWriter{
		st:    st,
		w:     writer,
		types: types,
	}

	if opts != nil {
		if opts.HeaderRow {
			writer.Write(titles)
		}
	}

	return wr
}

// TODO - holy shit dis so bad. fix
func terribleHackToGetHeaderRowAndTypes(st *dataset.Structure) ([]string, []string, error) {
	sch := st.Schema
	if itemObj, ok := sch["items"].(map[string]interface{}); ok {
		if itemArr, ok := itemObj["items"].([]interface{}); ok {
			titles := make([]string, len(itemArr))
			types := make([]string, len(itemArr))
			for i, f := range itemArr {
				if field, ok := f.(map[string]interface{}); ok {
					if title, ok := field["title"].(string); ok {
						titles[i] = title
					}

					if ts, ok := field["type"].(string); ok {
						types[i] = ts
					} else if ta, ok := field["type"].([]interface{}); ok && len(ta) > 0 {
						if st, ok := ta[0].(string); ok {
							types[i] = st
						} else {
							types[i] = "string"
						}
					} else {
						types[i] = "string"
					}
				}
			}
			return titles, types, nil
		}
	}
	log.Debug("that terrible hack to detect header row & types just failed")
	return nil, nil, fmt.Errorf("nope")
}

// Structure gives this writer's structure
func (w *CSVWriter) Structure() *dataset.Structure {
	return w.st
}

// WriteEntry writes one CSV record to the writer
func (w *CSVWriter) WriteEntry(ent Entry) error {
	if arr, ok := ent.Value.([]interface{}); ok {
		strs, err := encode(arr)
		if err != nil {
			log.Debug(err.Error())
			return fmt.Errorf("error encoding entry: %s", err.Error())
		}
		return w.w.Write(strs)
	}
	return fmt.Errorf("expected array value to write csv row. got: %v", ent)
}

// encode uses specified types from structure's schema to go values to strings
func encode(vs []interface{}) ([]string, error) {
	strings := make([]string, len(vs))

	for i, v := range vs {
		// vs[i] = str
		switch t := v.(type) {
		case string:
			strings[i] = t
		case int:
			strings[i] = strconv.Itoa(t)
		case int64:
			strings[i] = strconv.Itoa(int(t))
		case float64:
			strings[i] = strconv.FormatFloat(t, 'f', -1, 64)
		case []interface{}:
			if data, err := json.Marshal(t); err == nil {
				strings[i] = string(data)
			}
		case map[string]interface{}:
			if data, err := json.Marshal(t); err == nil {
				strings[i] = string(data)
			}
		case bool:
			if t {
				strings[i] = "true"
			} else {
				strings[i] = "false"
			}
		case nil:
			strings[i] = ""
		}
	}

	return strings, nil
}

// Close finalizes the writer, indicating no more records
// will be written
func (w *CSVWriter) Close() error {
	w.w.Flush()
	return nil
}
