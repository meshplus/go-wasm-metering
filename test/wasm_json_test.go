package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/meshplus/go-wasm-metering/json2wasm"
	"github.com/meshplus/go-wasm-metering/tool"
	"github.com/meshplus/go-wasm-metering/wasm2json"
	"github.com/stretchr/testify/assert"
)

var expectedJson = []tool.JSON{
	{
		"name":    "preramble",
		"magic":   []byte{0, 97, 115, 109},
		"version": []byte{1, 0, 0, 0},
	},
	{
		"name":         "custom",
		"section_name": "a custom section",
		"custom":       "this is the payload",
	},
}

func TestCustomSection(t *testing.T) {
	wasm, err := ioutil.ReadFile(path.Join("testdata", "customSection.wasm"))
	assert.Nil(t, err)
	jsonObj, err := wasm2json.Wasm2Json(wasm)
	assert.Nil(t, err)
	assert.Equal(t, true, assert.ObjectsAreEqual(expectedJson, jsonObj))

	generatedWasm, err := json2wasm.Json2Wasm(jsonObj)
	assert.Nil(t, err)
	assert.Equal(t, 0, bytes.Compare(generatedWasm, wasm))
	assert.Equal(t, true, assert.ObjectsAreEqual(expectedJson, jsonObj))
}

//func readWasmModule(path string) ([]tool.JSON, error) {
//	var jsonArr []tool.JSON
//	jsonData, err := ioutil.ReadFile(path)
//	if err != nil {
//		return nil, err
//	}
//	if err = json.Unmarshal(jsonData, &jsonArr); err != nil {
//		return nil, err
//	}
//	return jsonArr, nil
//}

func TestBasicTest(t *testing.T) {
	dirName := path.Join("testdata", "wasm")
	dir, err := ioutil.ReadDir(dirName)
	assert.Nil(t, err)
	failed := 0
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}

		wasm, err := ioutil.ReadFile(path.Join(dirName, fi.Name()))
		assert.Nil(t, err)

		jsonObj, err := wasm2json.Wasm2Json(wasm)
		assert.Nil(t, err)
		wasmBin, err := json2wasm.Json2Wasm(jsonObj)
		assert.Nil(t, err)

		if !assert.Equal(t, 0, bytes.Compare(wasm, wasmBin)) {
			failed += 1
			fmt.Println(fi.Name())
			fmt.Printf("%#v\n", jsonObj)
		}
	}

	fmt.Printf("total failed case %d\n", failed)
}

//func TestBasicTest1(t *testing.T) {
//	dirName := path.Join("testdata", "json")
//	dir, err := ioutil.ReadDir(dirName)
//	assert.Nil(t, err)
//	for _, fi := range dir {
//		if fi.IsDir() {
//			continue
//		}
//
//		jsonObj, err := readWasmModule(path.Join(dirName, fi.Name()))
//		assert.Nil(t, err)
//		fmt.Printf("%#v\n", jsonObj)
//
//		wasm, err := json2wasm.Json2Wasm(jsonObj)
//		assert.Nil(t, err)
//		jsonObj2, err := wasm2json.Wasm2Json(wasm)
//		assert.Nil(t, err)
//		fmt.Println(fi.Name())
//		fmt.Printf("%#v\n", jsonObj2)
//
//		assert.Equal(t, true, assert.ObjectsAreEqual(jsonObj, jsonObj2))
//	}
//}

func TestText2Json(t *testing.T) {
	text := "i32.const 32 drop"
	json := tool.Text2Json(text)

	expected := []tool.JSON{
		{
			"name":       "const",
			"returns":    "i32",
			"immediates": "32",
		}, {
			"name": "drop",
		},
	}

	assert.Equal(t, true, assert.ObjectsAreEqual(expected, json))

	text = "br_table 0 0 0 0 i64.const 24"
	json = tool.Text2Json(text)

	expected = []tool.JSON{
		{
			"name":       "br_table",
			"immediates": []string{"0", "0", "0", "0"},
		}, {
			"returns":    "i64",
			"name":       "const",
			"immediates": "24",
		},
	}

	assert.Equal(t, true, assert.ObjectsAreEqual(expected, json))

	text = "call_indirect 1 i64.const 24"
	json = tool.Text2Json(text)

	expected = []tool.JSON{
		{
			"name": "call_indirect",
			"immediates": tool.JSON{
				"index":    "1",
				"reserved": 0,
			},
		}, {
			"returns":    "i64",
			"name":       "const",
			"immediates": "24",
		},
	}

	assert.Equal(t, true, assert.ObjectsAreEqual(expected, json))

	text = "i32.load 0 1 i64.const 24"
	json = tool.Text2Json(text)

	expected = []tool.JSON{
		{
			"name":    "load",
			"returns": "i32",
			"immediates": tool.JSON{
				"flags":  "0",
				"offset": "1",
			},
		}, {
			"returns":    "i64",
			"name":       "const",
			"immediates": "24",
		},
	}

	assert.Equal(t, true, assert.ObjectsAreEqual(expected, json))
}
