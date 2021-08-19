package json2wasm

import (
	"fmt"

	"github.com/meshplus/go-wasm-metering/tool"
)

type typeGenerator struct{}

func (t typeGenerator) Function(num uint32, stream *tool.Stream) error {
	if _, err := tool.EncodeULEB128(num, stream); err != nil {
		return fmt.Errorf("type generator funtcion: %w", err)
	}
	return nil
}

func (typeGenerator) Table(table tool.Table, stream *tool.Stream) error {
	if _, err := stream.Write([]byte{J2W_LANGUAGE_TYPES[table.ElementType]}); err != nil {
		return fmt.Errorf("type generator table: %w", err)
	}

	if err := typeGen.Memory(table.Limits, stream); err != nil {
		return fmt.Errorf("type generator table: %w", err)
	}

	return nil
}

// Generates a [`global_type`](https://github.com/WebAssembly/design/blob/master/BinaryEncoding.md#global_type)
func (typeGenerator) Global(global tool.Global, stream *tool.Stream) error {
	if err := stream.WriteByte(J2W_LANGUAGE_TYPES[global.ContentType]); err != nil {
		return fmt.Errorf("type generator global: %w", err)
	}
	if err := stream.WriteByte(global.Mutability); err != nil {
		return fmt.Errorf("type generator global: %w", err)
	}
	return nil
}

// Generates a [resizable_limits](https://github.com/WebAssembly/design/blob/master/BinaryEncoding.md#resizable_limits)
func (typeGenerator) Memory(mem tool.MemLimits, stream *tool.Stream) error {
	if mem.Maximum != nil {
		if _, err := tool.EncodeULEB128(1, stream); err != nil {
			return fmt.Errorf("type generator memory: %w", err)
		}
		if _, err := tool.EncodeULEB128(mem.Intial, stream); err != nil {
			return fmt.Errorf("type generator memory: %w", err)
		}
		if _, err := tool.EncodeULEB128(mem.Maximum.(uint32), stream); err != nil {
			return fmt.Errorf("type generator memory: %w", err)
		}
	} else {
		if _, err := tool.EncodeULEB128(0, stream); err != nil {
			return fmt.Errorf("type generator memory: %w", err)
		}
		if _, err := tool.EncodeULEB128(mem.Intial, stream); err != nil {
			return fmt.Errorf("type generator memory: %w", err)
		}
	}
	return nil
}

func (typeGenerator) InitExpr(op tool.OP, stream *tool.Stream) error {
	if _, err := GenerateOP(op, stream); err != nil {
		return fmt.Errorf("type generator InitExpr: %w", err)
	}
	if _, err := GenerateOP(tool.OP{
		Name: "end",
		Type: "void",
	}, stream); err != nil {
		return fmt.Errorf("type generator InitExpr: %w", err)
	}

	return nil
}
