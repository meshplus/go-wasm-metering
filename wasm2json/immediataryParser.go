package wasm2json

import "github.com/meshplus/go-wasm-metering/tool"

type immediataryParser struct{}

func (immediataryParser) Varuint1(stream *tool.Stream) (int8, error) {
	b, err := stream.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(b), nil
}

func (immediataryParser) Varuint32(stream *tool.Stream) (uint32, error) {
	ret, err := tool.DecodeULEB128(stream)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (immediataryParser) Varint32(stream *tool.Stream) (int32, error) {
	ret, err := tool.DecodeSLEB128(stream)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (immediataryParser) Varint64(stream *tool.Stream) (int64, error) {
	ret, err := tool.DecodeSLEB128(stream)
	if err != nil {
		return 0, err
	}
	return int64(ret), nil
}

func (immediataryParser) Uint32(stream *tool.Stream) []byte {
	return stream.Read(4)
}

func (immediataryParser) Uint64(stream *tool.Stream) []byte {
	return stream.Read(8)
}

func (immediataryParser) BlockType(stream *tool.Stream) (string, error) {
	ret, err := stream.ReadByte()
	if err != nil {
		return "", err
	}
	return W2J_LANGUAGE_TYPES[ret], nil
}

func (immediataryParser) BrTable(stream *tool.Stream) (tool.JSON, error) {
	jsonObj := make(tool.JSON)
	targets := []uint32{}

	num, err := tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	for i := uint32(0); i < num; i++ {
		target, err := tool.DecodeULEB128(stream)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	jsonObj["targets"] = targets
	jsonObj["default_target"], err = tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	return jsonObj, nil
}

func (immediataryParser) CallIndirect(stream *tool.Stream) (tool.JSON, error) {
	jsonObj := make(tool.JSON)
	var err error
	jsonObj["index"], err = tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	jsonObj["reserved"], err = stream.ReadByte()
	if err != nil {
		return nil, err
	}
	return jsonObj, nil
}

func (immediataryParser) MemoryImmediate(stream *tool.Stream) (tool.JSON, error) {
	jsonObj := make(tool.JSON)
	var err error
	jsonObj["flags"], err = tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	jsonObj["offset"], err = tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	return jsonObj, nil
}
