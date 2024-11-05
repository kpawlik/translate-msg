package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

const LIMIT = 0
const DEBUG = false


type Files struct {
	InDir  string
	Files  []string
	OutDir string
	Lang   string
}


var (
	cnt = 0
	coreTest = Files{
		InDir:  `E:\repos\IQGeo-7.0\WebApps\myworldapp\public\locales\en`,
		OutDir: `E:\repos\IQGeo-7.0\WebApps\myworldapp\public\locales\nl`,
		Files: []string{
			`tmp.msg`,
		},
		Lang: "nl",
	}
	core = Files{
		InDir:  `E:\repos\IQGeo-7.0\WebApps\myworldapp\public\locales\en`,
		OutDir: `E:\repos\IQGeo-7.0\WebApps\myworldapp\public\locales\nl`,
		Files: []string{
			// `myw.client.msg`,
			// `myw.config.msg`,
			// `myw.system.settings.msg`,
			`myw.install.msg`,
		},
		Lang: "nl",
	}
	survey = Files{
		InDir:  `E:\repos\IQGeo_IS_v7.0.5.2\survey\public\locales\en`,
		OutDir: `E:\repos\IQGeo_IS_v7.0.5.2\survey\public\locales\nl`,
		Files: []string{
			`install.msg`,
			// `survey.config.msg`,
			// `survey.msg`,
		},
		Lang: "nl",
	}
	common = Files{
		InDir:  `E:\repos\IQGeo_IS_v7.0.5.2\mywapp_common\public\locales\en`,
		OutDir: `E:\repos\IQGeo_IS_v7.0.5.2\mywapp_common\public\locales\nl`,
		Files: []string{
			`mywapp_common.msg`,
		},
		Lang: "nl",
	}
	anywhere = Files{
		InDir: `E:\Projects\Stedin\Anywhere-message-files\anywhere\en`,
		OutDir: `E:\Projects\Stedin\Anywhere-message-files\anywhere\nl`,
		Files: []string{
			`myw.app.msg`,
		},
		Lang: "nl",
	}
	missing = Files{
		InDir: `C:\kpa-home\GoogleDrive\dev\GO\translate-msg\translations\en`,
		OutDir: `C:\kpa-home\GoogleDrive\dev\GO\translate-msg\translations\nl`,
		Files: []string{
			`missing.msg`,
		},
		Lang: "nl",
	}
	// modules = []Files{core, survey, common}
	modules = []Files{missing}
)


type OrderedMap struct{
	Map map[string]interface{}
	Keys []string
}

// Create a new OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Map:    make(map[string]interface{}),
		Keys: []string{},
	}
}

func (om *OrderedMap) ParseObject(dec *json.Decoder) (err error) {
	var t json.Token
	var value interface{}
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		value, err = HandleDelim(t, dec)
		if err != nil {
			return err
		}
		om.Map[key] = value
		om.Keys = append(om.Keys, key)
	}
		t, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}


// this implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, om)
func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expect JSON object open with '{'")
	}

	err = om.ParseObject(dec)
	if err != nil {
		return err
	}

	t, err = dec.Token()
	if err != io.EOF {
		return fmt.Errorf("expect end of JSON object but got more token: %T: %v or err: %v", t, t, err)
	}

	return nil
}


// this implements type json.Marshaler interface, so can be called in json.Marshal(om)
func (om *OrderedMap) MarshalJSON() (res []byte, err error) {
	lines := make([]string, len(om.Keys))
	for i, key := range om.Keys {
		var b []byte
		b, err = json.Marshal(om.Map[key])
		if err != nil {
			return
		}
		lines[i] = fmt.Sprintf("\"%s\": %s", key, string(b))
	}
	res = append(res, '{')
	res = append(res, []byte(strings.Join(lines, ","))...)
	res = append(res, '}')
	return
}

func ParseArray(dec *json.Decoder) (arr []interface{}, err error) {
	var t json.Token
	arr = make([]interface{}, 0)
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return
		}

		var value interface{}
		value, err = HandleDelim(t, dec)
		if err != nil {
			return
		}
		arr = append(arr, value)
	}
	t, err = dec.Token()
	if err != nil {
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		err = fmt.Errorf("expect JSON array close with ']'")
		return
	}

	return
}




func HandleDelim(t json.Token, dec *json.Decoder) (res interface{}, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewOrderedMap()
			err = om2.ParseObject(dec)
			if err != nil {
				return
			}
			return om2, nil
		case '[':
			var value []interface{}
			value, err = ParseArray(dec)
			if err != nil {
				return
			}
			return value, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter: %q", delim)
		}
	}
	return t, nil
}




func processJson(source *OrderedMap, lang string, client *translate.Client) (target *OrderedMap, err error) {
	target = NewOrderedMap()
	for _, namespace := range source.Keys {
		items := source.Map[namespace]
		targetNs := NewOrderedMap()
		target.Map[namespace] = targetNs
		target.Keys = append(target.Keys, namespace)
		itemsMap, _ := items.(*OrderedMap)
		for _, key := range itemsMap.Keys {
			targetNs.Keys = append(targetNs.Keys, key)
			text := itemsMap.Map[key]
			if LIMIT > 0 && cnt > LIMIT {
				break
			}
			switch val := text.(type) {
			case []any:
				targetNs.Map[key] = translateArray(val, lang, client)
			case string:
				translated := translateStr(val, lang, client)
				targetNs.Map[key] = translated
			}
		}
	}
	return
}


func main() {
	var (
		bytes  []byte
		err    error
		client *translate.Client
		target *OrderedMap
		
	
	)
	ctx := context.Background()
	client, err = translate.NewClient(ctx)
	if err != nil {
		log.Fatalf("Connect to translate client error, %v", err)
	}
	defer client.Close()
	source := NewOrderedMap()
	for _, module := range modules {
		log.Printf("Processing module %s\n", module.InDir)
		for _, file := range module.Files {
			log.Printf("File start: %s\n", file)
			filePath := filepath.Join(module.InDir, file)
			if bytes, err = os.ReadFile(filePath); err != nil {
				log.Fatal(fmt.Errorf("error reading file %s, %v", filePath, err))
			}
			if err := json.Unmarshal(bytes, &source); err != nil {
				log.Fatal(fmt.Errorf("error unmarshal file %s, %v", filePath, err))
			}
			if target, err = processJson(source, module.Lang, client); err != nil {
				log.Fatal(fmt.Errorf("error processing file %s, %v", filePath, err))
			}
			if bytes, err = json.MarshalIndent(target, "", "  "); err != nil{
				log.Fatal(fmt.Errorf("error marshal file %s, %v", filePath, err))
			}
			if err = os.MkdirAll(module.OutDir, os.ModeDir); err != nil{
				log.Fatalf("Error creating dir %s, %v", module.OutDir, err)
			}
			outPath := filepath.Join(module.OutDir, file)
			if err = os.WriteFile(outPath, bytes, os.ModePerm); err != nil{
				log.Fatal(fmt.Errorf("error writing file %s, %v", filePath, err))
			}
			log.Printf("File end: %s\n", file)
		}
	}
	fmt.Printf("Processed messages %d\n", cnt)
}




func mockTranslateStr(t string) string {
	return t
}

func translateStr(original string, lang string, client *translate.Client) string {
	cnt ++
	if DEBUG{
		return mockTranslateStr(original)
	}
	translated, err := translateText(lang, original, client)
	if err != nil {
		log.Fatal(err)
	}
	translated = replacePlaceholders(original, translated)
	translated = html.UnescapeString(translated)
	translated = html.EscapeString(translated)
	translated = strings.ReplaceAll(translated, "&#39;", "'")
	return translated
}

func translateArray(arr []any, lang string, client *translate.Client) []string {
	results := make([]string, len(arr))
	for i, text := range arr {
		value, _ := text.(string)
		translated := translateStr(value, lang, client)
		results[i] = translated
	}
	return results
}


func translateText(targetLanguage, text string, client *translate.Client) (string, error) {
	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %w", err)
	}
	resp, err := client.Translate(context.Background(), []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("translate: %w", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
}

func replacePlaceholders(original, translated string) string {
	var (
		result string
	)
	result = translated
	re := regexp.MustCompile(`(__\w+__)`)
	oPh := re.FindAllString(original, -1)
	tPh := re.FindAllString(translated, -1)
	if len(oPh) == len(tPh){
		for i, _ := range oPh {
			result = strings.Replace(result, tPh[i], oPh[i], 1)
		}
	}
	return result
}
