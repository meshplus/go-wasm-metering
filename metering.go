package go_wasm_metering

import (
	"github.com/meshplus/go-wasm-metering/json2wasm"
	"github.com/meshplus/go-wasm-metering/tool"
	"github.com/meshplus/go-wasm-metering/wasm2json"
	"github.com/sirupsen/logrus"
)

const (
	defaultModuleStr = "metering"
	defaultFieldStr  = "usegas"
	defaultMeterType = "i64"
	DefaultCost      = uint64(0)
)

type Options struct {
	CostTable tool.JSON // path of cost table file.
	ModuleStr string    // the import string for metering function.
	FieldStr  string    // the field string for the metering function.
	MeterType string    // the register type that is used to meter. Can be `i64`, `i32`, `f64`, `f32`.
}

// MeterWASM injects metering into WebAssembly binary code.
// This func is the real exported function used by outer callers.
func MeterWASM(wasm []byte, opts *Options, logger logrus.FieldLogger) ([]byte, uint64, error) {
	logger.WithFields(logrus.Fields{}).Info("MeterWASM is starting")
	// 1. covert wasm to json
	module, err := wasm2json.Wasm2Json(wasm)
	if err != nil {
		return nil, 0, err
	}
	logger.WithFields(logrus.Fields{}).Info("Wasm 2 json end")
	logger.WithFields(logrus.Fields{
		"ruleCode": module,
	}).Debug("Wasm code before metering")

	// 2. metering
	if opts == nil {
		opts = &Options{}
	}
	metering, err := newMetring(*opts, logger)
	if err != nil {
		return nil, 0, err
	}
	module, gasCost, err := metering.MeterJSON(module)
	if err != nil {
		return nil, 0, err
	}
	logger.WithFields(logrus.Fields{}).Info("Meter json end")

	// 3. covert json to wasm
	meteredWasm, err := json2wasm.Json2Wasm(module)
	if err != nil {
		return nil, 0, err
	}
	logger.WithFields(logrus.Fields{}).Info("Json 2 wasm end")

	return meteredWasm, gasCost, nil
}

func newMetring(opts Options, logger logrus.FieldLogger) (*Metering, error) {
	// set defaults.
	if opts.CostTable == nil {
		opts.CostTable = DefaultCostTable
	}

	if opts.ModuleStr == "" {
		opts.ModuleStr = defaultModuleStr
	}

	if opts.FieldStr == "" {
		opts.FieldStr = defaultFieldStr
	}

	if opts.MeterType == "" {
		opts.MeterType = defaultMeterType
	}

	return &Metering{
		Opts:   opts,
		Logger: logger,
	}, nil
}
