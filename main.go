package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
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
			// `mywapp_common.msg`,
		},
		Lang: "nl",
	}
	modules = []Files{core, survey, common}
	// modules = []Files{coreTest}
)


func main() {
	var (
		bytes  []byte
		err    error
		source map[string]map[string]interface{}
		target map[string]map[string]interface{}
		client *translate.Client
	)
	ctx := context.Background()
	client, err = translate.NewClient(ctx)
	if err != nil {
		log.Fatalf("Connect to translate client error, %v", err)
	}
	defer client.Close()
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



func processJson(source map[string]map[string]interface{}, lang string, client *translate.Client) (target map[string]map[string]interface{}, err error) {
	target = make(map[string]map[string]interface{})
	for namespace, items := range source {
		target[namespace] = make(map[string]interface{})
		for key, text := range items {
			if LIMIT > 0 && cnt > LIMIT {
				break
			}
			switch val := text.(type) {
			case []any:
				target[namespace][key] = translateArray(val, lang, client)
			case string:
				translated := translateStr(val, lang, client)
				target[namespace][key] = translated
			}
		}
	}
	return
}

// func test() {
// 	const IN_FILE = `tst-en.msg`
// 	const OUT_FILE = `tst-nl.msg`

// 	var (
// 		bytes []byte
// 		err   error
// 	)
// 	bytes, err = os.ReadFile(IN_FILE)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var source map[string]map[string]interface{}
// 	target := make(map[string]map[string]interface{})

// 	json.Unmarshal(bytes, &source)
// 	cnt := 0
// 	for namespace, items := range source {
// 		_ = namespace
// 		target[namespace] = make(map[string]interface{})
// 		for key, text := range items {
// 			if LIMIT > 0 && cnt > LIMIT {
// 				break
// 			}
// 			switch val := text.(type) {
// 			case []any:
// 				target[namespace][key] = translateArray(val, "nl")
// 			case string:
// 				// translated := translateStr(t)
// 				translated := mockTranslateStrFromFile(key, val)
// 				translated = replacePlaceholders(val, translated)
// 				target[namespace][key] = translated
// 			}
// 			cnt++
// 		}

// 	}
// 	// bytes, err = json.Marshal(target)
// 	// if err != nil{
// 	// 	log.Fatal(err)
// 	// }
// 	//os.WriteFile(OUT_FILE, bytes, os.ModePerm)
// 	print(target)

// }

// func mockTranslateStrFromFile(key, text string) string {
// 	cnt++
// 	var (
// 		bytes []byte
// 		err   error
// 	)
// 	bytes, err = os.ReadFile(OUT_FILE)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var source map[string]map[string]interface{}

// 	if err = json.Unmarshal(bytes, &source); err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, v := range source {
// 		for kk, vv := range v {
// 			if kk == key {
// 				return vv.(string)
// 			}
// 		}
// 	}
// 	return ""
// }

func mockTranslateStr(t string) string {
	return t
}

func translateStr(original string, lang string, client *translate.Client) string {
	cnt ++
	if DEBUG{
		return mockTranslateStr(original)
	}
	// fmt.Println(1, original)
	translated, err := translateText(lang, original, client)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(2, translated)
	translated = replacePlaceholders(original, translated)
	// fmt.Println(3, translated)
	translated = html.UnescapeString(translated)
	// fmt.Println(4, translated)
	translated = html.EscapeString(translated)
	translated = strings.ReplaceAll(translated, "&#39;", "'")
	// fmt.Println(5, translated)
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

// func print(obj map[string]map[string]interface{}) {
// 	for k, v := range obj {
// 		fmt.Println(k)
// 		for kk, vv := range v {
// 			fmt.Printf("   %s: %v\n", kk, vv)
// 		}
// 	}
// }

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
	// fmt.Println("---")
	// fmt.Println(original)
	// fmt.Println(translated)
	result = translated
	re := regexp.MustCompile(`(__\w+__)`)
	oPh := re.FindAllString(original, -1)
	tPh := re.FindAllString(translated, -1)
	if len(oPh) == len(tPh){
		for i, _ := range oPh {
			result = strings.Replace(result, tPh[i], oPh[i], 1)
		}
	}
	// fmt.Println(result)
	return result
}
