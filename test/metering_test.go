package test

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	metering "github.com/meshplus/go-wasm-metering"
	"github.com/meshplus/go-wasm-metering/wasm2json"
	"github.com/stretchr/testify/assert"
)

const (
	defaultModuleStr = "metering"
	defaultFieldStr  = "usegas"
	defaultMeterType = "i64"
	defaultCost      = uint64(0)
)

func TestMeter(t *testing.T) {
	// 1. in
	inWasm, err := ioutil.ReadFile(path.Join("testdata", "in", "wasm", "ledger_test_gc.wasm"))
	assert.Nil(t, err)

	meteredWasm, _, err := metering.MeterWASM(inWasm, &metering.Options{
		CostTable: metering.DefaultCostTable,
	})
	assert.Nil(t, err)

	// 2. out
	err = ioutil.WriteFile(path.Join("testdata", "out", "wasm", "ledger_test_gc-meter.wasm"), meteredWasm, 0644)
	assert.Nil(t, err)

	//// 3. json compare
	//meteredJson, err := wasm2json.Wasm2Json(meteredWasm)
	//assert.Nil(t, err)
	//inJson, err := wasm2json.Wasm2Json(inWasm)
	//assert.Nil(t, err)
	//
	//fmt.Printf("%#v\n%#v\n", meteredJson, inJson)
	//fmt.Printf("gas cost: %d\n", gasCost)
	//
	//fmt.Println("=================")
	//for i, sec := range meteredJson {
	//	fmt.Printf("%#v\n%#v\n\n", sec["entries"], inJson[i]["entries"])
	//}
}

func TestBasicMeteringTests(t *testing.T) {
	dirName := path.Join("testdata", "in")
	dir, err := ioutil.ReadDir(path.Join(dirName, "wasm"))
	assert.Nil(t, err)
	for _, file := range dir {
		// read wasm json.
		wasm, err := ioutil.ReadFile(path.Join(dirName, "wasm", file.Name()))
		assert.Nil(t, err)

		module, err := wasm2json.Wasm2Json(wasm)
		assert.Nil(t, err)

		// read cost table json.
		metering := metering.Metering{
			Opts: metering.Options{
				CostTable: metering.DefaultCostTable,
				ModuleStr: defaultModuleStr,
				FieldStr:  defaultFieldStr,
				MeterType: defaultMeterType,
			},
		}
		//fmt.Printf("%s %#v\n", file.Name(), module)
		meteredModule, _, err := metering.MeterJSON(module)
		if err != nil {
			assert.Equal(t, "basic+import.wasm", file.Name())
			continue
		}
		//fmt.Printf("%s old %#v\n", file.Name(), meteredModule)
		//fmt.Printf("Gas: %v\n", gasCost)

		expectedWasm, err := ioutil.ReadFile(path.Join("testdata", "expected-out", "wasm", file.Name()))
		assert.Nil(t, err)
		expectedJson, err := wasm2json.Wasm2Json(expectedWasm)
		assert.Nil(t, err)
		//fmt.Printf("%s exp %#v\n", file.Name(), expectedJson)

		if !assert.Equal(t, true, assert.ObjectsAreEqual(meteredModule, expectedJson)) {
			fmt.Printf("file name %s\n", file.Name())
			fmt.Printf("%#v\n%#v\n", meteredModule, expectedJson)
		}
	}
}
