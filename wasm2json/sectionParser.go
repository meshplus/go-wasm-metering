package wasm2json

import "github.com/meshplus/go-wasm-metering/tool"

type sectionParser struct{}

func (sectionParser) Custom(stream *tool.Stream, header tool.SectionHeader) (tool.CustomSec, error) {
	sec := tool.CustomSec{Name: "custom"}

	// create a new stream to read.
	section := tool.NewStream(stream.Read(int(header.Size)))
	nameLen, err := tool.DecodeULEB128(section)
	if err != nil {
		return tool.CustomSec{}, err
	}
	name := section.Read(int(nameLen))
	sec.SectionName = string(name)

	var custom interface{}
	switch string(name) {
	case "name":
		custom, err = cparser.CustomNames(section)
		if err != nil {
			return tool.CustomSec{}, err
		}
	default:
		custom = section.String()
	}

	sec.Custom = custom

	return sec, nil
}

func (sectionParser) Type(stream *tool.Stream) (tool.TypeSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.TypeSec{}, err
	}
	typSec := tool.TypeSec{
		Name:    "type",
		Entries: []tool.TypeEntry{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		typ, err := stream.ReadByte()
		if err != nil {
			return tool.TypeSec{}, err
		}
		entry := tool.TypeEntry{
			Form:   W2J_LANGUAGE_TYPES[typ],
			Params: []string{},
		}

		paramCount, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.TypeSec{}, err
		}

		// parse the entries.
		for j := uint32(0); j < paramCount; j++ {
			typ, err := stream.ReadByte()
			if err != nil {
				return tool.TypeSec{}, err
			}
			entry.Params = append(entry.Params, W2J_LANGUAGE_TYPES[typ])
		}

		numOfReturns, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.TypeSec{}, err
		}

		for j := uint32(0); j < numOfReturns; j++ {
			typ, err := stream.ReadByte()
			if err != nil {
				return tool.TypeSec{}, err
			}
			entry.Returns = append(entry.Returns, W2J_LANGUAGE_TYPES[typ])
		}

		typSec.Entries = append(typSec.Entries, entry)
	}

	return typSec, nil
}

func (s sectionParser) Import(stream *tool.Stream) (tool.ImportSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.ImportSec{}, err
	}
	importSec := tool.ImportSec{
		Name:    "import",
		Entries: []tool.ImportEntry{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		moduleLen, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ImportSec{}, err
		}
		moduleStr := stream.Read(int(moduleLen))

		fieldLen, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ImportSec{}, err
		}
		fieldStr := stream.Read(int(fieldLen))

		kind, err := stream.ReadByte()
		if err != nil {
			return tool.ImportSec{}, err
		}
		externalKind := W2J_EXTERNAL_KIND[kind]
		var returned interface{}
		switch externalKind {
		case "function":
			returned, err = tParser.Function(stream)
			if err != nil {
				return tool.ImportSec{}, err
			}
		case "table":
			returned, err = tParser.Table(stream)
			if err != nil {
				return tool.ImportSec{}, err
			}
		case "memory":
			returned, err = tParser.Memory(stream)
			if err != nil {
				return tool.ImportSec{}, err
			}
		case "global":
			returned, err = tParser.Global(stream)
			if err != nil {
				return tool.ImportSec{}, err
			}
		}

		entry := tool.ImportEntry{
			ModuleStr: string(moduleStr),
			FieldStr:  string(fieldStr),
			Kind:      externalKind,
			Type:      returned,
		}

		importSec.Entries = append(importSec.Entries, entry)
	}

	return importSec, nil
}

func (sectionParser) Function(stream *tool.Stream) (tool.FuncSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.FuncSec{}, err
	}
	funcSec := tool.FuncSec{
		Name:    "function",
		Entries: []uint32{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		entry, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.FuncSec{}, err
		}
		funcSec.Entries = append(funcSec.Entries, entry)
	}
	return funcSec, nil
}

func (s sectionParser) Table(stream *tool.Stream) (tool.TableSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.TableSec{}, err
	}
	tableSec := tool.TableSec{
		Name:    "table",
		Entries: []tool.Table{},
	}

	// parse table_type.
	for i := uint32(0); i < numberOfEntries; i++ {
		entry, err := tParser.Table(stream)
		if err != nil {
			return tool.TableSec{}, err
		}
		tableSec.Entries = append(tableSec.Entries, entry)
	}

	return tableSec, nil
}

func (sectionParser) Memory(stream *tool.Stream) (tool.MemSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.MemSec{}, err
	}
	memSec := tool.MemSec{
		Name:    "memory",
		Entries: []tool.MemLimits{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		entry, err := tParser.Memory(stream)
		if err != nil {
			return tool.MemSec{}, err
		}
		memSec.Entries = append(memSec.Entries, entry)
	}
	return memSec, nil
}

func (sectionParser) Global(stream *tool.Stream) (tool.GlobalSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.GlobalSec{}, err
	}
	globalSec := tool.GlobalSec{
		Name:    "global",
		Entries: []tool.GlobalEntry{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		t, err := tParser.Global(stream)
		if err != nil {
			return tool.GlobalSec{}, err
		}
		i, err := tParser.InitExpr(stream)
		if err != nil {
			return tool.GlobalSec{}, err
		}
		entry := tool.GlobalEntry{
			Type: t,
			Init: i,
		}

		globalSec.Entries = append(globalSec.Entries, entry)
	}

	return globalSec, nil
}

func (sectionParser) Export(stream *tool.Stream) (tool.ExportSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.ExportSec{}, err
	}
	exportSec := tool.ExportSec{
		Name:    "export",
		Entries: []tool.ExportEntry{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		strLength, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ExportSec{}, err
		}
		fieldStr := string(stream.Read(int(strLength)))
		kind, err := stream.ReadByte()
		if err != nil {
			return tool.ExportSec{}, err
		}
		index, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ExportSec{}, err
		}

		entry := tool.ExportEntry{
			FieldStr: fieldStr,
			Kind:     W2J_EXTERNAL_KIND[kind],
			Index:    index,
		}

		exportSec.Entries = append(exportSec.Entries, entry)
	}

	return exportSec, nil
}

func (sectionParser) Start(stream *tool.Stream) (tool.StartSec, error) {
	index, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.StartSec{}, err
	}
	startSec := tool.StartSec{
		Name:  "start",
		Index: index,
	}
	return startSec, nil
}

func (sectionParser) Element(stream *tool.Stream) (tool.ElementSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.ElementSec{}, err
	}
	elSec := tool.ElementSec{
		Name:    "element",
		Entries: []tool.ElementEntry{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		entry := tool.ElementEntry{}
		entry.Index, err = tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ElementSec{}, err
		}
		entry.Offset, err = tParser.InitExpr(stream)
		if err != nil {
			return tool.ElementSec{}, err
		}

		numElem, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.ElementSec{}, err
		}
		for j := uint32(0); j < numElem; j++ {
			elem, err := tool.DecodeULEB128(stream)
			if err != nil {
				return tool.ElementSec{}, err
			}
			entry.Elements = append(entry.Elements, elem)
		}

		elSec.Entries = append(elSec.Entries, entry)
	}

	return elSec, nil
}

func (sectionParser) Code(stream *tool.Stream) (tool.CodeSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.CodeSec{}, err
	}
	codeSec := tool.CodeSec{
		Name:    "code",
		Entries: []tool.CodeBody{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		codeBody := tool.CodeBody{
			Locals: []tool.LocalEntry{},
			Code:   []tool.OP{},
		}

		bodySize, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.CodeSec{}, err
		}
		endBytes := stream.BytesRead + int(bodySize)

		// parse locals
		localCount, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.CodeSec{}, err
		}
		for j := uint32(0); j < localCount; j++ {
			local := tool.LocalEntry{}
			local.Count, err = tool.DecodeULEB128(stream)
			if err != nil {
				return tool.CodeSec{}, err
			}
			typ, err := stream.ReadByte()
			if err != nil {
				return tool.CodeSec{}, err
			}
			local.Type = W2J_LANGUAGE_TYPES[typ]
			codeBody.Locals = append(codeBody.Locals, local)
		}

		// parse code
		for stream.BytesRead < endBytes {
			op, err := ParseOp(stream)
			if err != nil {
				return tool.CodeSec{}, err
			}
			codeBody.Code = append(codeBody.Code, op)
		}

		codeSec.Entries = append(codeSec.Entries, codeBody)
	}

	return codeSec, nil
}

func (sectionParser) Data(stream *tool.Stream) (tool.DataSec, error) {
	numberOfEntries, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.DataSec{}, err
	}
	dataSec := tool.DataSec{
		Name:    "data",
		Entries: []tool.DataSegment{},
	}

	for i := uint32(0); i < numberOfEntries; i++ {
		entry := tool.DataSegment{}
		entry.Index, err = tool.DecodeULEB128(stream)
		if err != nil {
			return tool.DataSec{}, err
		}
		entry.Offset, err = tParser.InitExpr(stream)
		if err != nil {
			return tool.DataSec{}, err
		}
		segmentSize, err := tool.DecodeULEB128(stream)
		if err != nil {
			return tool.DataSec{}, err
		}
		entry.Data = append([]byte{}, stream.Read(int(segmentSize))...)

		dataSec.Entries = append(dataSec.Entries, entry)
	}

	return dataSec, nil
}

func (sectionParser) DataCount(stream *tool.Stream) (tool.DataCountSec, error) {
	count, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.DataCountSec{}, err
	}
	dataCountSec := tool.DataCountSec{
		Name:  "data_count",
		Count: count,
	}

	return dataCountSec, nil
}
