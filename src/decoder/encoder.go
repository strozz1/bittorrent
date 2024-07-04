package decoder

import (
	"fmt"
	"slices"
	"strconv"
)

func Encode(m map[string]any) ([]byte, error) {
    return encode_dict(m)
}
func encode_dict(m map[string]any) ([]byte, error) {
	word := []byte{}
	keys := get_keys(m)
	for _, key := range keys {
		value := m[key]
		k_len := fmt.Sprintf("%d", len(key))

		word = append(word, []byte(k_len)...)
		word = append(word, []byte(key)...) // len:word

		switch value.(type) {
		case int:
			num, _ := value.(int)
			encoded := encode_int(num)
			word = append(word, []byte(encoded)...)
		case []byte:
			s, _ := value.([]byte)
			length := []byte(string(len(s)) + ":")
			encoded := []byte{}
			encoded = append(encoded, length...)
			encoded = append(encoded, value.([]byte)...)

			word = append(word, encoded...)
		case string:
			s, _ := value.(string)
			encoded := fmt.Sprintf("%d:%s", len(s), s)
			word = append(word, []byte(encoded)...)

		case map[string]any:
			res, _ := encode_dict(value.(map[string]any))
			word = append(word, res...)

		default:
			return nil, fmt.Errorf("Error encoding")

		}

	}
	return word, nil

}

func get_keys(m map[string]any) []string {
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func encode_int(num int) []byte {
	encoded := "i" + strconv.Itoa(num) + "e"
	return []byte(encoded)
}


