package go_wasm_metering

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/meshplus/go-wasm-metering/json2wasm"
	"github.com/meshplus/go-wasm-metering/tool"
	"github.com/sirupsen/logrus"
)

type Metering struct {
	Opts   Options
	Logger logrus.FieldLogger
}

var (
	branchOps = map[string]struct{}{
		"grow_memory": {},
		"end":         {},
		"br":          {},
		"br_table":    {},
		"br_if":       {},
		"if":          {},
		"else":        {},
		"return":      {},
		"loop":        {},
	}
)

// meterJSON injects metering into a JSON output of Wasm2Json.
func (m *Metering) MeterJSON(module []tool.JSON) ([]tool.JSON, uint64, error) {
	// 1. add necessary `type` and `import` sections if and only if they don't exist.
	if m.findSection(module, "type") == nil {
		module = m.createSection(module, "type")
	}
	if m.findSection(module, "import") == nil {
		module = m.createSection(module, "import")
	}

	// 2. prepare
	importEntry := tool.ImportEntry{
		ModuleStr: m.Opts.ModuleStr,
		FieldStr:  m.Opts.FieldStr,
		Kind:      "function",
	}

	importType := tool.TypeEntry{
		Form:   "func",
		Params: []string{m.Opts.MeterType},
	}

	importCusName := tool.NameAssoc{
		NameStr: fmt.Sprintf("%s.%s", m.Opts.ModuleStr, m.Opts.FieldStr),
	}

	var (
		typeModule     tool.JSON
		functionModule tool.JSON
		funcIndex      int
		newModule      = make([]tool.JSON, len(module))
		gasCost        uint64
	)

	copy(newModule, module)

	// 3. Insert module by module
	for _, section := range newModule {
		sectionName, exist := section["name"]
		if !exist {
			continue
		}
		switch sectionName.(string) {
		case "type":
			var entries []tool.TypeEntry
			ientries, exist := section["entries"]
			if exist {
				entries = ientries.([]tool.TypeEntry)
			}
			importEntry.Type = uint32(len(entries))
			entries = append(entries, importType)
			section["entries"] = entries

			// save for use for the code section.
			typeModule = section
		case "function":
			// save for use for the code section.
			functionModule = section
		case "import":
			var entries []tool.ImportEntry
			ientries, exist := section["entries"]
			if exist {
				entries = ientries.([]tool.ImportEntry)
			}
			for _, entry := range entries {
				if entry.ModuleStr == m.Opts.ModuleStr && entry.FieldStr == m.Opts.FieldStr {
					return nil, 0, fmt.Errorf("importing metering function is not allowed")
				}

				if entry.Kind == "function" {
					funcIndex += 1
				}
			}
			// append the metering import.
			section["entries"] = append(entries, importEntry)
		case "export":
			var entries []tool.ExportEntry
			ientries, exist := section["entries"]
			if exist {
				entries = ientries.([]tool.ExportEntry)
			}
			for i, entry := range entries {
				if entry.Kind == "function" && entry.Index >= uint32(funcIndex) {
					entries[i].Index = entry.Index + 1
				}
			}
		case "element":
			var entries []tool.ElementEntry
			ientries, exist := section["entries"]
			if exist {
				entries = ientries.([]tool.ElementEntry)
			}
			for i, entry := range entries {
				// remap element indices.
				newElements := make([]uint32, 0, len(entry.Elements))
				for _, el := range entry.Elements {
					if el >= uint32(funcIndex) {
						el += 1
					}
					newElements = append(newElements, el)
				}
				entries[i].Elements = newElements
			}
		case "start":
			index := section["index"].(uint32)
			if index >= uint32(funcIndex) {
				index += 1
			}
			section["index"] = index
		case "code":
			entries := section["entries"].([]tool.CodeBody)
			funcEntries := functionModule["entries"].([]uint32)
			typEntries := typeModule["entries"].([]tool.TypeEntry)
			for i, entry := range entries {
				typeIndex := funcEntries[i]
				typ := typEntries[typeIndex]
				cost := m.getCost(typ, m.Opts.CostTable["type"].(tool.JSON), DefaultCost)

				entry, cost, err := m.meterCodeEntry(entry, m.Opts.CostTable["code"].(tool.JSON), m.Opts.MeterType, funcIndex, cost)
				m.Logger.WithFields(logrus.Fields{
					"code": entry,
				}).Debug("meter entry")
				if err != nil {
					return nil, 0, fmt.Errorf("meterCodeEntry error: %v", err)
				}
				gasCost += cost
				entries[i] = entry
			}
		case "custom":
			var customNames []tool.CustomName
			sectionName, exist := section["section_name"]
			if exist && sectionName == "name" {
				iCustomNames, exist1 := section["custom"]
				if exist1 {
					customNames = iCustomNames.([]tool.CustomName)
					for i, cusName := range customNames {
						switch cusName.Kind {
						case "function":
							names := cusName.Names.([]tool.NameAssoc)
							newNames := []tool.NameAssoc{}
							for _, functionName := range names {
								if functionName.Index >= uint32(funcIndex) {
									if functionName.Index == uint32(funcIndex) {
										importCusName.Index = uint32(funcIndex)
										newNames = append(newNames, importCusName)
									}
									functionName.Index++
								}
								newNames = append(newNames, functionName)
							}
							customNames[i].Names = newNames
						case "local":
							names := cusName.Names.([]tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌)
							newNames := []tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌{}
							for _, functionLocals := range names {
								if functionLocals.Index >= uint32(funcIndex) {
									functionLocals.Index++
								}
							}
							customNames[i].Names = newNames
						}
					}
				}
			}

		}
	}
	return newModule, gasCost, nil
}

func (m *Metering) findSection(module []tool.JSON, sectionName string) tool.JSON {
	for _, section := range module {
		if name, exist := section["name"]; exist {
			if name.(string) == sectionName {
				return section
			}
		}
	}
	return nil
}

func (m *Metering) createSection(module []tool.JSON, sectionName string) []tool.JSON {
	newSectionId := json2wasm.J2W_SECTION_IDS[sectionName]
	for i, section := range module {
		name, exist := section["name"]
		if exist {
			secId, exist := json2wasm.J2W_SECTION_IDS[name.(string)]
			if exist && secId > 0 && newSectionId < secId {
				rest := append([]tool.JSON{}, module[i:]...)
				// insert the section at pos `i`
				module = append(module[:i], tool.JSON{
					"name": sectionName,
				})
				module = append(module, rest...)
				break
			}
		}
	}
	return module
}

// meter code json========================================================================================
// getCost returns the cost of an operation for the entry in a section from the cost table.
func (m *Metering) getCost(j interface{}, costTable tool.JSON, defaultCost uint64) (cost uint64) {
	if dc, exist := costTable["DEFAULT"]; exist {
		defaultCost = uint64(dc.(int))
	}
	rval := reflect.ValueOf(j)
	kind := rval.Type().Kind()
	if kind == reflect.Slice {
		for i := 0; i < rval.Len(); i++ {
			cost += m.getCost(rval.Index(i).Interface(), costTable, 0)
		}
	} else if kind == reflect.Struct {
		rtype := rval.Type()
		for i := 0; i < rval.NumField(); i++ {
			rv := rval.Field(i)
			propCost, exist := costTable[tool.Lcfirst(rtype.Field(i).Name)]
			if exist {
				cost += m.getCost(rv.Interface(), propCost.(tool.JSON), defaultCost)
			} else {
				cost += defaultCost
			}
		}
	} else if kind == reflect.String {
		key := j.(string)
		if key == "" {
			return 0
		}
		c, exist := costTable[key]
		if exist {
			cost = uint64(c.(int))
		} else {
			cost = defaultCost
		}
	} else {
		cost = defaultCost
	}
	//fmt.Printf("json %#v cost %v\n", j, cost)
	return
}

// meterCodeEntry meters a single code entry (see tool.CodeBody).
func (m *Metering) meterCodeEntry(entry tool.CodeBody, costTable tool.JSON, meterType string, meterFuncIndex int, cost uint64) (tool.CodeBody, uint64, error) {
	getImmediateFromOP := func(name, opType string) string {
		var immediatesKey string
		if name == "const" {
			immediatesKey = opType
		} else {
			immediatesKey = name
		}
		return tool.OP_IMMEDIATES[immediatesKey]
	}

	meteringStatement := func(cost uint64, meteringImportIndex int) (ops []tool.OP) {
		opsJson := tool.Text2Json(fmt.Sprintf("%s.const %v call %v", meterType, cost, meteringImportIndex))
		for _, op := range opsJson {

			oop := tool.OP{
				Name: op["name"].(string),
			}

			// convert immediates.
			imm := getImmediateFromOP(oop.Name, meterType)
			if imm != "" {
				opImm := op["immediates"]
				switch imm {
				case "varuint1":
					imme, _ := strconv.ParseInt(opImm.(string), 10, 8)
					oop.Immediates = int8(imme)
				case "varuint32":
					imme, _ := strconv.ParseUint(opImm.(string), 10, 32)
					oop.Immediates = uint32(imme)
				case "varint32":
					imme, _ := strconv.ParseInt(opImm.(string), 10, 32)
					oop.Immediates = int32(imme)
				case "varint64":
					imme, _ := strconv.ParseInt(opImm.(string), 10, 64)
					oop.Immediates = int64(imme)
				case "uint32":
					oop.Immediates = opImm.([]byte)
				case "uint64":
					oop.Immediates = opImm.([]byte)
				case "block_type":
					oop.Immediates = opImm.(string)
				case "br_table", "call_indirect", "memory_immediate":
					oop.Immediates = opImm.(tool.JSON)
				}
			}

			if rt, ok := op["returns"]; ok {
				oop.ReturnType = rt.(string)
			}

			if rt, ok := op["type"]; ok {
				oop.Type = rt.(string)
			}

			ops = append(ops, oop)
		}

		return
	}

	remapOp := func(op *tool.OP, funcIndex int) {
		if op.Name == "call" {
			switch imm := op.Immediates.(type) {
			case string:
				rv, _ := strconv.ParseInt(imm, 10, 64)
				if rv >= int64(funcIndex) {
					rv += 1
					op.Immediates = strconv.FormatInt(rv, 10)
				}
			case uint32:
				if imm >= uint32(funcIndex) {
					imm += 1
					op.Immediates = imm
				}
			default:
				panic(fmt.Sprintf("invalid immediates type: %v", imm))
			}

		}
	}

	meterTheMeteringStatement := func() uint64 {
		code := meteringStatement(0, meterFuncIndex)
		// sum the operations cost
		sum := uint64(0)
		for _, op := range code {
			sum += m.getCost(op.Name, costTable["code"].(tool.JSON), DefaultCost)
		}
		return sum
	}

	var (
		meteringCost = meterTheMeteringStatement()
		code         = make([]tool.OP, len(entry.Code))
		meteredCode  []tool.OP
	)
	//fmt.Printf("meter the meter cost %d\n", meteringCost)

	// create a code copy.
	copy(code, entry.Code)

	cost += m.getCost(entry.Locals, costTable["locals"].(tool.JSON), DefaultCost)
	sum := uint64(0)

	for len(code) > 0 {
		i := 0

		// meter a segment of wasm code.
		for {
			op := &code[i]
			remapOp(op, meterFuncIndex)
			cost += m.getCost(code[i].Name, costTable["code"].(tool.JSON), DefaultCost)
			i += 1
			if _, exist := branchOps[op.Name]; exist {
				break
			}
			if i >= len(code) {
				return tool.CodeBody{}, 0, fmt.Errorf("illegal code(the lack of branch) : %v", code)
			}
		}

		// add the metering statement.
		if cost != 0 {
			// add the cost of metering
			cost += meteringCost
			ops := meteringStatement(cost, meterFuncIndex)
			meteredCode = append(meteredCode, ops...)
		}
		sum += cost

		meteredCode = append(meteredCode, code[:i]...)
		code = code[i:]
		cost = 0
	}

	entry.Code = meteredCode
	return entry, sum, nil
}
