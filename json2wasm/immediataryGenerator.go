package json2wasm

import (
	"fmt"

	"github.com/meshplus/go-wasm-metering/tool"
)

type immediataryGenerator struct{}

func (immediataryGenerator) Varuint1(j int8, stream *tool.Stream) (*tool.Stream, error) {
	if err := stream.WriteByte(byte(j)); err != nil {
		return nil, fmt.Errorf("immediatary generator Varuint1: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) Varuint32(j uint32, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeULEB128(uint32(j), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator Varuint32: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) Varint32(j int32, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeSLEB128(int32(j), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator Varint32: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) Varint64(j int64, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeSLEB128(int32(j), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator Varint64: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) Uint32(j []byte, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := stream.Write(j); err != nil {
		return nil, fmt.Errorf("immediatary generator Uint32: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) Uint64(j []byte, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := stream.Write(j); err != nil {
		return nil, fmt.Errorf("immediatary generator Uint64: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) BlockType(j string, stream *tool.Stream) (*tool.Stream, error) {
	if err := stream.WriteByte(J2W_LANGUAGE_TYPES[j]); err != nil {
		return nil, fmt.Errorf("immediatary generator BlockType: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) BrTable(j tool.JSON, stream *tool.Stream) (*tool.Stream, error) {
	targets := j["targets"].([]uint32)
	if _, err := tool.EncodeULEB128(uint32(len(targets)), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator BrTable: %w", err)
	}

	for _, target := range targets {
		if _, err := tool.EncodeULEB128(target, stream); err != nil {
			return nil, fmt.Errorf("immediatary generator BrTable: %w", err)
		}
	}
	if _, err := tool.EncodeULEB128(j["default_target"].(uint32), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator Varuint1: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) CallIndirect(j tool.JSON, stream *tool.Stream) (*tool.Stream, error) {
	index := j["index"]
	if _, err := tool.EncodeULEB128(index.(uint32), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator CallIndirect: %w", err)
	}
	if err := stream.WriteByte(j["reserved"].(byte)); err != nil {
		return nil, fmt.Errorf("immediatary generator CallIndirect: %w", err)
	}
	return stream, nil
}

func (immediataryGenerator) MemoryImmediate(j tool.JSON, stream *tool.Stream) (*tool.Stream, error) {
	if _, err := tool.EncodeULEB128(j["flags"].(uint32), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator MemoryImmediate: %w", err)
	}
	if _, err := tool.EncodeULEB128(j["offset"].(uint32), stream); err != nil {
		return nil, fmt.Errorf("immediatary generator MemoryImmediate: %w", err)
	}
	return stream, nil
}
