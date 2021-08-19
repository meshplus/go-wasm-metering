package wasm2json

import "github.com/meshplus/go-wasm-metering/tool"

type typeParser struct{}

func (typeParser) Function(stream *tool.Stream) (uint32, error) {
	return tool.DecodeULEB128(stream)
}

func (t typeParser) Table(stream *tool.Stream) (tool.Table, error) {
	typ, err := stream.ReadByte()
	if err != nil {
		return tool.Table{}, err
	}
	limits, err := t.Memory(stream)
	if err != nil {
		return tool.Table{}, err
	}
	return tool.Table{
		ElementType: W2J_LANGUAGE_TYPES[typ],
		Limits:      limits,
	}, nil
}

func (typeParser) Global(stream *tool.Stream) (tool.Global, error) {
	typ, err := stream.ReadByte()
	if err != nil {
		return tool.Global{}, err
	}
	mutability, err := stream.ReadByte()
	if err != nil {
		return tool.Global{}, err
	}
	return tool.Global{
		ContentType: W2J_LANGUAGE_TYPES[typ],
		Mutability:  mutability,
	}, nil
}

func (typeParser) Memory(stream *tool.Stream) (tool.MemLimits, error) {
	flags, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.MemLimits{}, err
	}
	intial, err := tool.DecodeULEB128(stream)
	if err != nil {
		return tool.MemLimits{}, err
	}
	limits := tool.MemLimits{
		Flags:  flags,
		Intial: intial,
	}
	if flags == 1 {
		limits.Maximum, err = tool.DecodeULEB128(stream)
		if err != nil {
			return tool.MemLimits{}, err
		}
	}
	return limits, nil
}

func (typeParser) InitExpr(stream *tool.Stream) (tool.OP, error) {
	op, err := ParseOp(stream)
	if err != nil {
		return tool.OP{}, err
	}
	_, err = stream.ReadByte() // skip the `end`
	if err != nil {
		return tool.OP{}, err
	}
	return op, nil
}
