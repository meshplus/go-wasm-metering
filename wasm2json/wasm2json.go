package wasm2json

import (
	"strings"

	"github.com/meshplus/go-wasm-metering/tool"
)

var (
	immeParser = immediataryParser{}
	tParser    = typeParser{}
	secParser  = sectionParser{}
	cparser    = customParser{}
)

// Wasm2Json convert the wasm binary to a JSON array output.
func Wasm2Json(buf []byte) ([]tool.JSON, error) {
	stream := tool.NewStream(buf)
	preramble := ParsePreramble(stream)
	resJson := []tool.JSON{preramble}

	for stream.Len() != 0 {
		header, err := ParseSectionHeader(stream)
		if err != nil {
			return nil, err
		}
		jsonObj := make(tool.JSON)
		switch header.Name {
		case "custom":
			rsec, err := secParser.Custom(stream, header)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["section_name"] = rsec.SectionName
			jsonObj["custom"] = rsec.Custom
		case "type":
			rsec, err := secParser.Type(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "import":
			rsec, err := secParser.Import(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "function":
			rsec, err := secParser.Function(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "table":
			rsec, err := secParser.Table(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "memory":
			rsec, err := secParser.Memory(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "global":
			rsec, err := secParser.Global(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "export":
			rsec, err := secParser.Export(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "start":
			rsec, err := secParser.Start(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["index"] = rsec.Index
		case "element":
			rsec, err := secParser.Element(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "code":
			rsec, err := secParser.Code(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "data":
			rsec, err := secParser.Data(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["entries"] = rsec.Entries
		case "data count":
			rsec, err := secParser.DataCount(stream)
			if err != nil {
				return nil, err
			}
			jsonObj["name"] = rsec.Name
			jsonObj["count"] = rsec.Count
		}

		resJson = append(resJson, jsonObj)
	}

	return resJson, nil
}

func ParsePreramble(stream *tool.Stream) tool.JSON {
	magic := stream.Read(4)
	version := stream.Read(4)

	jsonObj := make(tool.JSON)
	jsonObj["name"] = "preramble"
	jsonObj["magic"] = magic
	jsonObj["version"] = version

	return jsonObj
}

func ParseSectionHeader(stream *tool.Stream) (tool.SectionHeader, error) {
	id, err := stream.ReadByte()
	if err != nil {
		return tool.SectionHeader{}, err
	}

	size, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.SectionHeader{}, err
	}

	return tool.SectionHeader{
		Id:   id,
		Name: W2J_SECTION_IDS[id],
		Size: size,
	}, nil
}

func ParseOp(stream *tool.Stream) (tool.OP, error) {
	finalOP := tool.OP{}
	op, err := stream.ReadByte()
	if err != nil {
		return tool.OP{}, err
	}
	fullName := strings.Split(W2J_OPCODES[op], ".")
	var (
		typ           = fullName[0]
		name          string
		immediatesKey string
	)

	if len(fullName) < 2 {
		name = typ
	} else {
		name = fullName[1]
		finalOP.ReturnType = typ
	}

	finalOP.Name = name

	if name == "const" {
		immediatesKey = typ
	} else {
		immediatesKey = name
	}
	immediates, exist := tool.OP_IMMEDIATES[immediatesKey]
	if exist {
		var returned interface{}
		switch immediates {
		case "block_type":
			returned, err = immeParser.BlockType(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "call_indirect":
			returned, err = immeParser.CallIndirect(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "varuint32":
			returned, err = immeParser.Varuint32(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "varuint1":
			returned, err = immeParser.Varuint1(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "varint32":
			returned, err = immeParser.Varint32(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "varint64":
			returned, err = immeParser.Varint64(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "uint32":
			returned = immeParser.Uint32(stream)
		case "uint64":
			returned = immeParser.Uint64(stream)
		case "br_table":
			returned, err = immeParser.BrTable(stream)
			if err != nil {
				return tool.OP{}, err
			}
		case "memory_immediate":
			returned, err = immeParser.MemoryImmediate(stream)
			if err != nil {
				return tool.OP{}, err
			}
		}
		finalOP.Immediates = returned
	}

	return finalOP, nil
}
