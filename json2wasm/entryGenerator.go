package json2wasm

import (
	"fmt"

	"github.com/meshplus/go-wasm-metering/tool"
)

type entryGenerator struct{}

func (entryGenerator) Type(entry tool.TypeEntry, stream *tool.Stream) ([]byte, error) {
	// a single type entry binary encoded
	if err := stream.WriteByte(J2W_LANGUAGE_TYPES[entry.Form]); err != nil {
		return nil, fmt.Errorf("entry generator type: %w", err)
	}

	// number of parameters
	paramsLen := len(entry.Params)
	if _, err := tool.EncodeULEB128(uint32(paramsLen), stream); err != nil {
		return nil, fmt.Errorf("entry generator type: %w", err)
	}
	if paramsLen != 0 {
		paramsType := make([]byte, 0, paramsLen)
		for _, typ := range entry.Params {
			paramsType = append(paramsType, J2W_LANGUAGE_TYPES[typ])
		}
		if _, err := stream.Write(paramsType); err != nil {
			return nil, fmt.Errorf("entry generator type: %w", err)
		}
	}

	// number of return types
	returnsLen := len(entry.Returns)
	if _, err := tool.EncodeULEB128(uint32(returnsLen), stream); err != nil {
		return nil, fmt.Errorf("entry generator type: %w", err)
	}
	if returnsLen != 0 {
		returnsType := make([]byte, 0, returnsLen)
		for _, re := range entry.Returns {
			returnsType = append(returnsType, J2W_LANGUAGE_TYPES[re])
		}
		if _, err := stream.Write(returnsType); err != nil {
			return nil, fmt.Errorf("entry generator type: %w", err)
		}
	}

	return stream.Bytes(), nil
}

func (entryGenerator) Import(entry tool.ImportEntry, stream *tool.Stream) error {
	// write the module string
	moduleStr := entry.ModuleStr
	if _, err := tool.EncodeULEB128(uint32(len(moduleStr)), stream); err != nil {
		return fmt.Errorf("entry generator import: %w", err)
	}
	if _, err := stream.Write([]byte(moduleStr)); err != nil {
		return fmt.Errorf("entry generator import: %w", err)
	}
	// write the field string
	fieldStr := entry.FieldStr
	if _, err := tool.EncodeULEB128(uint32(len(fieldStr)), stream); err != nil {
		return fmt.Errorf("entry generator import: %w", err)
	}
	if _, err := stream.Write([]byte(fieldStr)); err != nil {
		return fmt.Errorf("entry generator import: %w", err)
	}

	if err := stream.WriteByte(J2W_EXTERNAL_KIND[entry.Kind]); err != nil {
		return fmt.Errorf("entry generator import: %w", err)
	}

	switch entry.Kind {
	case "function":
		if err := typeGen.Function(entry.Type.(uint32), stream); err != nil {
			return fmt.Errorf("entry generator import: %w", err)
		}
	case "table":
		if err := typeGen.Table(entry.Type.(tool.Table), stream); err != nil {
			return fmt.Errorf("entry generator import: %w", err)
		}
	case "memory":
		if err := typeGen.Memory(entry.Type.(tool.MemLimits), stream); err != nil {
			return fmt.Errorf("entry generator import: %w", err)
		}
	case "global":
		if err := typeGen.Global(entry.Type.(tool.Global), stream); err != nil {
			return fmt.Errorf("entry generator import: %w", err)
		}
	}

	return nil
}

func (entryGenerator) Function(entry uint32, stream *tool.Stream) ([]byte, error) {
	if _, err := tool.EncodeULEB128(entry, stream); err != nil {
		return nil, fmt.Errorf("entry generator function: %w", err)
	}
	return stream.Bytes(), nil
}

func (entryGenerator) Table(j tool.Table, stream *tool.Stream) error {
	if err := typeGen.Table(j, stream); err != nil {
		return fmt.Errorf("entry generator table: %w", err)
	}
	return nil
}

func (entryGenerator) Global(entry tool.GlobalEntry, stream *tool.Stream) (*tool.Stream, error) {

	if err := typeGen.Global(entry.Type, stream); err != nil {
		return nil, fmt.Errorf("entry generator global: %w", err)
	}
	if err := typeGen.InitExpr(entry.Init, stream); err != nil {
		return nil, fmt.Errorf("entry generator global: %w", err)
	}
	return stream, nil
}

func (entryGenerator) Memory(entry tool.MemLimits, stream *tool.Stream) error {
	if err := typeGen.Memory(entry, stream); err != nil {
		return fmt.Errorf("entry generator memory: %w", err)
	}

	return nil
}

func (entryGenerator) Export(entry tool.ExportEntry, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeULEB128(uint32(len(entry.FieldStr)), stream); err != nil {
		return nil, fmt.Errorf("entry generator export: %w", err)
	}
	if _, err := stream.Write([]byte(entry.FieldStr)); err != nil {
		return nil, fmt.Errorf("entry generator export: %w", err)
	}
	if err := stream.WriteByte(J2W_EXTERNAL_KIND[entry.Kind]); err != nil {
		return nil, fmt.Errorf("entry generator export: %w", err)
	}
	if _, err := tool.EncodeULEB128(uint32(entry.Index), stream); err != nil {
		return nil, fmt.Errorf("entry generator export: %w", err)
	}
	return stream, nil
}

func (entryGenerator) Element(entry tool.ElementEntry, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeULEB128(uint32(entry.Index), stream); err != nil {
		return nil, fmt.Errorf("entry generator element: %w", err)
	}
	if err := typeGen.InitExpr(entry.Offset, stream); err != nil {
		return nil, fmt.Errorf("entry generator element: %w", err)
	}
	if _, err := tool.EncodeULEB128(uint32(len(entry.Elements)), stream); err != nil {
		return nil, fmt.Errorf("entry generator element: %w", err)
	}
	for _, elem := range entry.Elements {
		if _, err := tool.EncodeULEB128(elem, stream); err != nil {
			return nil, fmt.Errorf("entry generator element: %w", err)
		}
	}
	return stream, nil
}

func (entryGenerator) Code(entry tool.CodeBody, stream *tool.Stream) (*tool.Stream, error) {
	codeStream := tool.NewStream(nil)
	// write the locals
	if _, err := tool.EncodeULEB128(uint32(len(entry.Locals)), codeStream); err != nil {
		return nil, fmt.Errorf("entry generator code: %w", err)
	}
	for _, local := range entry.Locals {
		if _, err := tool.EncodeULEB128(local.Count, codeStream); err != nil {
			return nil, fmt.Errorf("entry generator code: %w", err)
		}
		if err := codeStream.WriteByte(J2W_LANGUAGE_TYPES[local.Type]); err != nil {
			return nil, fmt.Errorf("entry generator code: %w", err)
		}
	}

	// write opcode
	for _, op := range entry.Code {
		if _, err := GenerateOP(op, codeStream); err != nil {
			return nil, fmt.Errorf("entry generator code: %w", err)
		}
	}

	if _, err := tool.EncodeULEB128(uint32(codeStream.BytesWrote), stream); err != nil {
		return nil, fmt.Errorf("entry generator code: %w", err)
	}
	if _, err := stream.Write(codeStream.Bytes()); err != nil {
		return nil, fmt.Errorf("entry generator code: %w", err)
	}
	return stream, nil
}

func (entryGenerator) Data(entry tool.DataSegment, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeULEB128(entry.Index, stream); err != nil {
		return nil, fmt.Errorf("entry generator data: %w", err)
	}
	if err := typeGen.InitExpr(entry.Offset, stream); err != nil {
		return nil, fmt.Errorf("entry generator data: %w", err)
	}
	if _, err := tool.EncodeULEB128(uint32(len(entry.Data)), stream); err != nil {
		return nil, fmt.Errorf("entry generator data: %w", err)
	}
	if _, err := stream.Write(entry.Data); err != nil {
		return nil, fmt.Errorf("entry generator data: %w", err)
	}
	return stream, nil
}
