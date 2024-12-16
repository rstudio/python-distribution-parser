package distributions

import (
	"fmt"
	"io"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/charmap"
)

func parse(r io.Reader) (*mail.Message, error) {
	return mail.ReadMessage(r)
}
func getHeaderValue(data *mail.Message, header string) string {
	return collapseLeadingWS(header, data.Header.Get(header))
}
func getAllHeaderValues(data *mail.Message, header string) []string {
	values := data.Header[header]
	headers := make([]string, 0)
	for _, value := range values {
		headers = append(headers, collapseLeadingWS(header, value))
	}
	return headers
}

func mustDecode(value interface{}) (string, error) {
	switch v := value.(type) {
	case []byte:
		// First try to decode as utf-8
		utfStr := string(v)
		for _, runeValue := range utfStr {
			if runeValue == 0xfffd { // unicode replacement character
				// if utf-8 fails, try decoding as latin1
				latin1Decoder := charmap.ISO8859_1.NewDecoder()
				decoded, err := latin1Decoder.String(utfStr)
				if err != nil {
					return "", err
				}
				return decoded, nil
			}
		}
		return utfStr, nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

func collapseLeadingWS(header, txt string) string {
	if strings.ToLower(header) == "description" || strings.ToLower(header) == "license" { // preserve newlines
		lines := strings.Split(strings.TrimSpace(txt), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "        ") { // 8 spaces
				lines[i] = line[8:]
			}
		}
		return strings.Join(lines, "\n")
	} else {
		lines := strings.FieldsFunc(txt, func(r rune) bool {
			return unicode.IsSpace(r) && r != '\n'
		})
		return strings.Join(lines, " ")
	}
}

type HeaderAttr struct {
	HeaderName string
	AttrName   string
	Multiple   bool
}

var HeaderAttrs1_0 = []HeaderAttr{ // PEP 241
	{"Metadata-Version", "metadata_version", false},
	{"Name", "name", false},
	{"Version", "version", false},
	{"Platform", "platform", true},
	{"Supported-Platform", "supported_platform", true},
	{"Summary", "summary", false},
	{"Description", "description", false},
	{"Keywords", "keywords", false},
	{"Home-Page", "home_page", false},
	{"Author", "author", false},
	{"Author-Email", "author_email", false},
	{"License", "license", false},
}

var HeaderAttrs1_1 = append(HeaderAttrs1_0, []HeaderAttr{ // PEP 314
	{"Classifier", "classifiers", true},
	{"Download-Url", "download_url", false},
	{"Requires", "requires", true},
	{"Provides", "provides", true},
	{"Obsoletes", "obsoletes", true},
}...)

var HeaderAttrs1_2 = append(HeaderAttrs1_1, []HeaderAttr{ // PEP 345
	{"Maintainer", "maintainer", false},
	{"Maintainer-Email", "maintainer_email", false},
	{"Requires-Python", "requires_python", false},
	{"Requires-External", "requires_external", true},
	{"Requires-Dist", "requires_dist", true},
	{"Provides-Dist", "provides_dist", true},
	{"Obsoletes-Dist", "obsoletes_dist", true},
	{"Project-Url", "project_urls", true},
}...)

var HeaderAttrs2_0 = HeaderAttrs1_2 //XXX PEP 426?

var HeaderAttrs2_1 = append(HeaderAttrs1_2, []HeaderAttr{ // PEP 566
	{"Provides-Extra", "provides_extra", true},
	{"Description-Content-Type", "description_content_type", false},
}...)

var HeaderAttrs2_2 = append(HeaderAttrs2_1, HeaderAttr{"Dynamic", "dynamic", true}) // PEP 643

var HeaderAttrs = map[string][]HeaderAttr{
	"1.0": HeaderAttrs1_0,
	"1.1": HeaderAttrs1_1,
	"1.2": HeaderAttrs1_2,
	"2.0": HeaderAttrs2_0,
	"2.1": HeaderAttrs2_1,
	"2.2": HeaderAttrs2_2,
}

type Distribution interface {
	ExtractMetadata() error
	Parse(data []byte) error

	// GetName is a helper method to get values
	GetName() string
	GetVersion() string
	GetPythonVersion() string

	// MetadataMap is used to return a map of all the metadata,
	// similar to how twine passes the metadata in a multipart
	// form request.
	MetadataMap() map[string][]string
}

type BaseDistribution struct {
	MetadataVersion string `json:"metadata_version"`
	// version 1.0
	Name               string   `json:"name"`
	Version            string   `json:"version"`
	Platforms          []string `json:"platform"`
	SupportedPlatforms []string `json:"supported_platform"`
	Summary            string   `json:"summary"`
	Description        string   `json:"description"`
	Keywords           string   `json:"keywords"`
	HomePage           string   `json:"home_page"`
	DownloadURL        string   `json:"download_url"`
	Author             string   `json:"author"`
	AuthorEmail        string   `json:"author_email"`
	License            string   `json:"license"`
	// version 1.1
	Classifiers []string `json:"classifiers"`
	Requires    []string `json:"requires"`
	Provides    []string `json:"provides"`
	Obsoletes   []string `json:"obsoletes"`
	// version 1.2
	Maintainer       string   `json:"maintainer"`
	MaintainerEmail  string   `json:"maintainer_email"`
	RequiresPython   string   `json:"requires_python"`
	RequiresExternal []string `json:"requires_external"`
	RequiresDist     []string `json:"requires_dist"`
	ProvidesDist     []string `json:"provides_dist"`
	ObsoletesDist    []string `json:"obsoletes_dist"`
	ProjectURLs      []string `json:"project_urls"`
	// version 2.1
	ProvidesExtras         []string `json:"provides_extra"`
	DescriptionContentType string   `json:"description_content_type"`
	// version 2.2
	Dynamic []string `json:"dynamic"`
}

func (bd *BaseDistribution) GetHeaderAttrs() ([]HeaderAttr, error) {
	ha, exists := HeaderAttrs[bd.MetadataVersion]
	if !exists {
		return []HeaderAttr{}, fmt.Errorf("header attributes for metadata version %s not found", bd.MetadataVersion)
	}

	return ha, nil
}

func (bd *BaseDistribution) Parse(data []byte) error {
	decodedData, err := mustDecode(data)
	if err != nil {
		return err
	}
	msg, err := parse(strings.NewReader(decodedData))
	if err != nil {
		return err
	}

	headerValue := getHeaderValue(msg, "Metadata-Version")
	if bd.MetadataVersion == "" && headerValue != "" {
		bd.MetadataVersion = headerValue
	}

	headerAttrs, err := bd.GetHeaderAttrs()
	if err != nil {
		return err
	}

	for _, headerAttr := range headerAttrs {
		if headerAttr.AttrName == "metadata_version" {
			continue
		}

		headerValues := getAllHeaderValues(msg, headerAttr.HeaderName)
		if len(headerValues) != 0 {
			if headerAttr.Multiple {
				err := bd.setJSONValue(headerAttr.AttrName, headerValues)
				if err != nil {
					return err
				}
			} else if headerValues[0] != "UNKNOWN" {
				err := bd.setJSONValue(headerAttr.AttrName, headerValues[0])
				if err != nil {
					return err
				}
			}
		}

	}

	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %v\n", err)

	}
	if body != nil {
		err := bd.setJSONValue("description", string(body))
		if err != nil {
			return err
		}
	}
	return nil
}

func (bd *BaseDistribution) GetName() string {
	return bd.Name
}

func (bd *BaseDistribution) GetVersion() string {
	return bd.Version
}

// TODO: remember to implement this for other distributions if they need something more specific (e.g. Wheels)
func (bd *BaseDistribution) GetPythonVersion() string {
	return ""
}

func (bd *BaseDistribution) setJSONValue(fieldName string, value interface{}) error {
	v := reflect.ValueOf(bd).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		jsonTag := fieldInfo.Tag.Get("json")
		if jsonTag == fieldName {
			fieldVal := v.Field(i)
			if !fieldVal.CanSet() {
				return fmt.Errorf("cannot set field %s", fieldName)
			}

			fieldType := fieldVal.Type()
			val := reflect.ValueOf(value)
			if val.Type().ConvertibleTo(fieldType) {
				fieldVal.Set(val.Convert(fieldType))
			} else {
				return fmt.Errorf("provided value type didn't match object field type")
			}
			return nil
		}
	}
	return fmt.Errorf("no such json field: %s in object", fieldName)
}

func StructToMap(input interface{}) map[string][]string {
	result := make(map[string][]string)

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := v.Type().Field(i).Name
		fieldValue := v.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			fieldName = jsonTag
		}

		if fieldValue.Kind() == reflect.Struct {
			subMap := StructToMap(fieldValue.Interface())
			for subKey, subValue := range subMap {
				result[subKey] = subValue
			}
		} else {
			switch val := fieldValue.Interface().(type) {
			case string:
				result[fieldName] = []string{val}
			case []string:
				result[fieldName] = val
			default:
				result[fieldName] = []string{fmt.Sprintf("%v", val)}
			}
		}
	}

	return result
}

// Convert an arbitrary string to a standard distribution name.
// Any runs of non-alphanumeric/. characters are replaced with a single '-'.
// Copied from pkg_resources.safe_name for compatibility with warehouse.
// See https://github.com/pypa/twine/issues/743.
func SafeName(name string) string {
	reg := regexp.MustCompile("[^A-Za-z0-9.]+")
	return reg.ReplaceAllString(name, "-")
}
