package dataset

import (
	"fmt"
)

// FormatConfig is the interface for data format configurations
type FormatConfig interface {
	// Format gives the data format being configured
	Format() DataFormat
	// map gives an object of configuration details
	Map() map[string]interface{}
}

// ParseFormatConfigMap returns a FormatConfig implementation for a given data format
// and options map, often used in decoding from recorded formats like, say, JSON
func ParseFormatConfigMap(f DataFormat, opts map[string]interface{}) (FormatConfig, error) {
	switch f {
	case CSVDataFormat:
		return NewCSVOptions(opts)
	case JSONDataFormat:
		return NewJSONOptions(opts)
	case XLSXDataFormat:
		return NewXLSXOptions(opts)
	default:
		return nil, fmt.Errorf("cannot parse configuration for format: %s", f.String())
	}
}

// NewCSVOptions creates a CSVOptions pointer from a map
func NewCSVOptions(opts map[string]interface{}) (*CSVOptions, error) {
	o := &CSVOptions{}
	if opts == nil {
		return o, nil
	}

	if opts["headerRow"] != nil {
		if headerRow, ok := opts["headerRow"].(bool); ok {
			o.HeaderRow = headerRow
		} else {
			return nil, fmt.Errorf("invalid headerRow value: %s", opts["headerRow"])
		}
	}

	if opts["lazyQuotes"] != nil {
		if lq, ok := opts["lazyQuotes"].(bool); ok {
			o.LazyQuotes = lq
		} else {
			return nil, fmt.Errorf("invalid lazyQuotes value: %s", opts["lazyQuotes"])
		}
	}

	if opts["separator"] != nil {
		if sep, ok := opts["separator"].(string); ok {
			if len(sep) != 1 {
				return nil, fmt.Errorf("separator must be a single character")
			}
			o.Separator = rune(sep[0])
		} else {
			return nil, fmt.Errorf("invalid separator value: %v", opts["separator"])
		}
	}

	if opts["variadicFields"] != nil {
		if vf, ok := opts["variadicFields"].(bool); ok {
			o.VariadicFields = vf
		} else {
			return nil, fmt.Errorf("invalid variadicFields value: %s", opts["variadicFields"])
		}
	}

	return o, nil
}

// CSVOptions specifies configuration details for csv files
// This'll expand in the future to interoperate with okfn csv spec
type CSVOptions struct {
	// HeaderRow specifies weather this csv file has a header row or not
	HeaderRow bool `json:"headerRow"`
	// If LazyQuotes is true, a quote may appear in an unquoted field and a
	// non-doubled quote may appear in a quoted field.
	LazyQuotes bool `json:"lazyQuotes"`
	// Separator is the field delimiter.
	// It is set to comma (',') by NewReader.
	// Comma must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	Separator rune `json:"separator,omitempty"`
	// VariadicFields sets permits records to have a variable number of fields
	// avoid using this
	VariadicFields bool `json:"variadicFields"`
}

// Format announces the CSV Data Format for the FormatConfig interface
func (*CSVOptions) Format() DataFormat {
	return CSVDataFormat
}

// Map returns a map[string]interface representation of the configuration
func (o *CSVOptions) Map() map[string]interface{} {
	if o == nil {
		return nil
	}
	opt := map[string]interface{}{}
	if o.HeaderRow {
		opt["headerRow"] = o.HeaderRow
	}
	if o.LazyQuotes {
		opt["lazyQuotes"] = o.LazyQuotes
	}
	if o.VariadicFields {
		opt["variadicFields"] = o.VariadicFields
	}
	if o.Separator != rune(0) {
		opt["separator"] = o.Separator
	}
	return opt
}

// NewJSONOptions creates a JSONOptions pointer from a map
func NewJSONOptions(opts map[string]interface{}) (*JSONOptions, error) {
	o := &JSONOptions{}
	if opts == nil {
		return o, nil
	}
	return o, nil
}

// JSONOptions specifies configuration details for json file format
type JSONOptions struct {
	// TODO:
	// Indent string
}

// Format announces the JSON Data Format for the FormatConfig interface
func (*JSONOptions) Format() DataFormat {
	return JSONDataFormat
}

// Map returns a map[string]interface representation of the configuration
func (o *JSONOptions) Map() map[string]interface{} {
	if o == nil {
		return nil
	}
	return map[string]interface{}{}
}

// XLSXOptions specifies configuraiton details for the xlsx file format
type XLSXOptions struct {
	SheetName string `json:"sheetName,omitempty"`
}

// NewXLSXOptions creates a XLSXOptions pointer from a map
func NewXLSXOptions(opts map[string]interface{}) (FormatConfig, error) {
	o := &XLSXOptions{}
	if opts == nil {
		return o, nil
	}

	if opts["sheetName"] != nil {
		if sheetName, ok := opts["sheetName"].(string); ok {
			o.SheetName = sheetName
		} else {
			return nil, fmt.Errorf("invalid sheetName value: %v", opts["sheetName"])
		}
	}

	return o, nil
}

// Format announces the XLSX data format for the FormatConfig interface
func (*XLSXOptions) Format() DataFormat {
	return XLSXDataFormat
}

// Map structures XLSXOptions as a map of string keys to values
func (o *XLSXOptions) Map() map[string]interface{} {
	if o == nil {
		return nil
	}
	opt := map[string]interface{}{}
	if o.SheetName != "" {
		opt["sheetName"] = o.SheetName
	}

	return opt
}
