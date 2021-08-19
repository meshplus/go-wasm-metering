package json2wasm

import (
	"fmt"

	"github.com/meshplus/go-wasm-metering/tool"
)

var (
	typeGen  = typeGenerator{}
	immeGen  = immediataryGenerator{}
	entryGen = entryGenerator{}
	cusGen   = customGenerator{}
)

// Json2Wasm converts a JSON array to wasm binary.
func Json2Wasm(j []tool.JSON) ([]byte, error) {
	stream := tool.NewStream(nil)
	preamble := j[0]
	if _, err := GeneratePreramble(preamble, stream); err != nil {
		return nil, fmt.Errorf("json 2 wasm error: %w", err)
	}
	rest := j[1:]
	for _, item := range rest {
		if _, err := GenerateSection(item, stream); err != nil {
			return nil, fmt.Errorf("json 2 wasm error: %w", err)
		}
	}

	return stream.Bytes(), nil
}

func GeneratePreramble(j tool.JSON, stream *tool.Stream) (*tool.Stream, error) {
	if stream == nil {
		stream = tool.NewStream(nil)
	}

	_, err := stream.Write(tool.Interface2Bytes(j["magic"]))
	if err != nil {
		return nil, fmt.Errorf("generate preramble error: %w", err)
	}

	_, err = stream.Write(tool.Interface2Bytes(j["version"]))
	if err != nil {
		return nil, fmt.Errorf("generate preramble error: %w", err)
	}

	return stream, nil
}

func GenerateOP(op tool.OP, stream *tool.Stream) (*tool.Stream, error) {
	if stream == nil {
		stream = tool.NewStream(nil)
	}

	name := op.Name
	if op.ReturnType != "" {
		name = op.ReturnType + "." + name
	}

	if err := stream.WriteByte(J2W_OPCODES[name]); err != nil {
		return nil, fmt.Errorf("generate op error: %w", err)
	}

	immediateKey := op.Name
	if immediateKey == "const" {
		immediateKey = op.ReturnType
	}
	immediates, exist := tool.OP_IMMEDIATES[immediateKey]
	if exist {
		switch immediates {
		case "block_type":
			if _, err := immeGen.BlockType(op.Immediates.(string), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "varuint32":
			if _, err := immeGen.Varuint32(op.Immediates.(uint32), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "varint32":
			if _, err := immeGen.Varint32(op.Immediates.(int32), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "varint64":
			if _, err := immeGen.Varint64(op.Immediates.(int64), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "varuint1":
			if _, err := immeGen.Varuint1(op.Immediates.(int8), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "uint32":
			if _, err := immeGen.Uint32(op.Immediates.([]byte), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "uint64":
			if _, err := immeGen.Uint64(op.Immediates.([]byte), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "call_indirect":
			if _, err := immeGen.CallIndirect(op.Immediates.(tool.JSON), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "memory_immediate":
			if _, err := immeGen.MemoryImmediate(op.Immediates.(tool.JSON), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		case "br_table":
			if _, err := immeGen.BrTable(op.Immediates.(tool.JSON), stream); err != nil {
				return nil, fmt.Errorf("generate op error: %w", err)
			}
		default:
			return nil, fmt.Errorf("generate preramble error: invalid op immediate: %s", immediates)
		}
	}
	return stream, nil
}

func GenerateSection(j tool.JSON, stream *tool.Stream) (*tool.Stream, error) {
	if stream == nil {
		stream = tool.NewStream(nil)
	}

	var name string
	nameinterf, exist := j["name"]
	if exist {
		name = nameinterf.(string)
	}
	payload := tool.NewStream(nil)
	err := stream.WriteByte(J2W_SECTION_IDS[name])
	if err != nil {
		return nil, fmt.Errorf("generate section error: %w", err)
	}

	if name == "custom" {
		sectionName := j["section_name"].(string)
		if _, err := tool.EncodeULEB128(uint32(len(sectionName)), payload); err != nil {
			return nil, fmt.Errorf("generate section error: %w", err)
		}
		_, err = payload.Write([]byte(sectionName))
		if err != nil {
			return nil, fmt.Errorf("generate section error: %w", err)
		}

		if sectionName == "name" {
			custom := j["custom"].([]tool.CustomName)
			if _, err := cusGen.CustomName(custom, payload); err != nil {
				return nil, fmt.Errorf("generate section error: %w", err)
			}
		} else {
			_, err = payload.Write([]byte(j["custom"].(string)))
			if err != nil {
				return nil, fmt.Errorf("generate section error: %w", err)
			}
		}
	} else if name == "start" {
		if _, err := tool.EncodeULEB128(j["index"].(uint32), payload); err != nil {
			return nil, fmt.Errorf("generate section error: %w", err)
		}
	} else {
		ientries, exist := j["entries"]
		if exist {
			//fmt.Printf("Gen %v\n", name)
			switch name {
			case "type":
				entries := ientries.([]tool.TypeEntry)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Type(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "import":
				entries := ientries.([]tool.ImportEntry)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if err := entryGen.Import(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "function":
				entries := ientries.([]uint32)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Function(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "table":
				entries := ientries.([]tool.Table)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if err := entryGen.Table(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "memory":
				entries := ientries.([]tool.MemLimits)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if err := entryGen.Memory(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "global":
				entries := ientries.([]tool.GlobalEntry)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Global(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "export":
				entries := ientries.([]tool.ExportEntry)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Export(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "element":
				entries := ientries.([]tool.ElementEntry)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Element(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "code":
				entries := ientries.([]tool.CodeBody)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Code(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			case "data":
				entries := ientries.([]tool.DataSegment)
				if _, err := tool.EncodeULEB128(uint32(len(entries)), payload); err != nil {
					return nil, fmt.Errorf("generate section error: %w", err)
				}
				for _, entry := range entries {
					if _, err := entryGen.Data(entry, payload); err != nil {
						return nil, fmt.Errorf("generate section error: %w", err)
					}
				}
			default:
				panic(fmt.Sprintf("invalid section name: %s", name))
			}
		}
	}

	// write the size of the payload.
	if _, err := tool.EncodeULEB128(uint32(payload.BytesWrote), stream); err != nil {
		return nil, fmt.Errorf("generate section error: %w", err)
	}
	_, err = stream.Write(payload.Bytes())
	if err != nil {
		return nil, fmt.Errorf("generate section error: %w", err)
	}
	return stream, nil
}
