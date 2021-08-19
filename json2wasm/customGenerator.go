package json2wasm

import (
	"fmt"

	"github.com/meshplus/go-wasm-metering/tool"
)

type customGenerator struct{}

func (customGenerator) CustomName(custom []tool.CustomName, payload *tool.Stream) (*tool.Stream, error) {
	for _, cusName := range custom {
		if err := payload.WriteByte(J2W_CUSTOM_NAME_TYPES[cusName.Kind]); err != nil {
			return nil, fmt.Errorf("custom generator name: %w", err)
		}

		subPayload := tool.NewStream(nil)
		switch cusName.Kind {
		case "module":
			if _, err := subPayload.Write([]byte(cusName.Names.(string))); err != nil {
				return nil, fmt.Errorf("custom generator name: %w", err)
			}
		case "function":
			funcNames := cusName.Names.([]tool.NameAssoc)
			if _, err := tool.EncodeULEB128(uint32(len(funcNames)), subPayload); err != nil {
				return nil, fmt.Errorf("custom generator name: %w", err)
			}
			for _, funcName := range funcNames {
				if err := subPayload.WriteByte(byte(funcName.Index)); err != nil {
					return nil, fmt.Errorf("custom generator name: %w", err)
				}
				if _, err := tool.EncodeULEB128(uint32(len(funcName.NameStr)), subPayload); err != nil {
					return nil, fmt.Errorf("custom generator name: %w", err)
				}
				if _, err := subPayload.Write([]byte(funcName.NameStr)); err != nil {
					return nil, fmt.Errorf("custom generator name: %w", err)
				}
			}
		case "local":
			functionLocals := cusName.Names.([]tool.I𝚗𝚍𝚒𝚛𝚎𝚌𝚝N𝚊𝚖𝚎A𝚜𝚜𝚘𝚌)
			if _, err := tool.EncodeULEB128(uint32(len(functionLocals)), subPayload); err != nil {
				return nil, fmt.Errorf("custom generator name: %w", err)
			}
			for _, funcLocal := range functionLocals {
				if err := subPayload.WriteByte(byte(funcLocal.Index)); err != nil {
					return nil, fmt.Errorf("custom generator name: %w", err)
				}
				for _, local := range funcLocal.NameMap {
					if err := subPayload.WriteByte(byte(local.Index)); err != nil {
						return nil, fmt.Errorf("custom generator name: %w", err)
					}
					if _, err := tool.EncodeULEB128(uint32(len(local.NameStr)), subPayload); err != nil {
						return nil, fmt.Errorf("custom generator name: %w", err)
					}
					if _, err := subPayload.Write([]byte(local.NameStr)); err != nil {
						return nil, fmt.Errorf("custom generator name: %w", err)
					}
				}
			}
		}

		if _, err := tool.EncodeULEB128(uint32(subPayload.BytesWrote), payload); err != nil {
			return nil, fmt.Errorf("custom generator name: %w", err)
		}
		if _, err := payload.Write(subPayload.Bytes()); err != nil {
			return nil, fmt.Errorf("custom generator name: %w", err)
		}
	}
	return payload, nil
}
