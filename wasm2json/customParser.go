package wasm2json

import "github.com/meshplus/go-wasm-metering/tool"

type customParser struct{}

func (customParser) ModuleName(stream *tool.Stream) (string, error) {
	bodySize, err := tool.DecodeULEB128(stream)
	if err != nil {
		return "", err
	}
	return string(stream.Read(int(bodySize))), nil
}

func (customParser) FunctionNames(stream *tool.Stream) ([]tool.NameAssoc, error) {
	bodySize, err := tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	endBytes := stream.BytesRead + int(bodySize)

	nameMap := []tool.NameAssoc{}

	for stream.BytesRead < endBytes {
		num, err := tool.DecodeULEB128(stream)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(num); i++ {
			index, err := stream.ReadByte()
			if err != nil {
				return nil, err
			}
			nameLen, err := tool.DecodeULEB128(stream)
			if err != nil {
				return nil, err
			}
			nameStr := stream.Read(int(nameLen))
			name := tool.NameAssoc{
				Index:   uint32(index),
				NameStr: string(nameStr),
			}
			nameMap = append(nameMap, name)
		}
	}

	return nameMap, nil
}

func (customParser) LocalNames(stream *tool.Stream) ([]tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌, error) {
	bodySize, err := tool.DecodeULEB128(stream)
	if err != nil {
		return nil, err
	}
	endBytes := stream.BytesRead + int(bodySize)

	var i𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌Map []tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌

	for stream.BytesRead < endBytes {
		num, err := tool.DecodeULEB128(stream)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(num); i++ {
			index, err := stream.ReadByte()
			if err != nil {
				return nil, err
			}
			inName := tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌{
				Index:   uint32(index),
				NameMap: []tool.NameAssoc{},
			}

			inNameNum, err := tool.DecodeULEB128(stream)
			if err != nil {
				return nil, err
			}
			for i := 0; i < int(inNameNum); i++ {
				index, err := stream.ReadByte()
				if err != nil {
					return nil, err
				}
				nameLen, err := tool.DecodeULEB128(stream)
				if err != nil {
					return nil, err
				}
				nameStr := stream.Read(int(nameLen))
				name := tool.NameAssoc{
					Index:   uint32(index),
					NameStr: string(nameStr),
				}
				inName.NameMap = append(inName.NameMap, name)
			}
			i𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌Map = append(i𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌Map, inName)
		}
	}

	return i𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌Map, err
}

func (customParser) CustomNames(stream *tool.Stream) ([]tool.CustomName, error) {
	cusNames := []tool.CustomName{}

	// parse name
	for stream.BytesRead < stream.Length {
		typ, err := stream.ReadByte()
		if err != nil {
			return nil, err
		}
		var returned interface{}

		switch typ {
		case 0x00:
			returned, err = cparser.ModuleName(stream)
			if err != nil {
				return nil, err
			}
		case 0x01:
			returned, err = cparser.FunctionNames(stream)
			if err != nil {
				return nil, err
			}
		case 0x02:
			returned, err = cparser.LocalNames(stream)
			if err != nil {
				return nil, err
			}
		}

		cusNames = append(cusNames, tool.CustomName{
			Kind:  W2J_CUSTOM_NAME_TYPES[typ],
			Names: returned,
		})
	}

	return cusNames, nil
}
