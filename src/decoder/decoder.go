package decoder

import (
	"fmt"
	"strconv"
	"unicode"
)

func Decode(bencode []byte) (any, error) {
	if bencode[0] == 'i' { // number
		return decode_int(string(bencode))
	} else if unicode.IsNumber(rune(bencode[0])) { // string
		return decode_string(bencode)
	} else if bencode[0] == 'l' { // list
		return decode_list(bencode)
	} else if bencode[0] == 'd' { // dict
		return decode_dict(bencode)
	}
	return nil, fmt.Errorf("Error parsing code")

}

func decode_bytes(bencode []byte) (any, error) {
	return decode_dict(bencode)
}
func decode_string(bencode []byte) ([]byte, error) {
	numbers := ""
	var colonIndex int

	if !(unicode.IsNumber(rune(bencode[0]))) { //if does not start with number
		return nil, fmt.Errorf("String must start with a number")
	}
	for i, letter := range bencode {
		if letter == ':' { // indicating end of number
			colonIndex = i
			break
		}
		if unicode.IsNumber(rune(bencode[0])) { // if number
			numbers += string(letter)
		}
	}
	length, _ := strconv.Atoi(numbers)
	if length+len(numbers)+1 != len(bencode) {
		return nil, fmt.Errorf("Length for string is not correct, %d, %d", length+len(numbers)+1, len(bencode))
	}

	return bencode[colonIndex+1 : colonIndex+1+length], nil
}

func decode_int(bencode string) (int, error) {
	if !(bencode[0] == 'i' && bencode[len(bencode)-1] == 'e') {
		return 0, fmt.Errorf("Number must start with i and end with e")
	}
	return strconv.Atoi(bencode[1 : len(bencode)-1])
}

func decode_dict(bencode []byte) (map[string]any, error) {
	if bencode[0] != 'd' {
		return nil, fmt.Errorf("Dictionary must start with d")
	}
	result, _ := dict_rec(bencode[1:])
	return result, nil
}

func dict_rec(bencode []byte) (map[string]any, []byte) {
	var dict = make(map[string]any)

	word := []byte{}
	key := ""
	var value any
	//iterate all letters
	for len(bencode) > 0 {
		char := bencode[0]
		bencode = bencode[1:] // -1 char

		if char == 'd' && len(word) == 0 { // if dict
			dict_res, code := dict_rec(bencode)
			bencode = code
			value = dict_res // set value to dict

		} else if char == 'e' && len(word) == 0 { //if end of dict

			return dict, bencode
		} else if char == 'l' && len(word) == 0 { //todo
			list_res, code := list_rec(bencode)
			bencode = code
			value = list_res // set value to dict

		} else { // string or number
			word = append(word, char)                       // add char to word
			if word[0] == 'i' && word[len(word)-1] == 'e' { // number
				number, _ := decode_int(string(word))
				value = number
			} else if unicode.IsDigit(rune(word[0])) {
				result, _ := decode_string(word)
				if result != nil {
					if key == "" {
						key = string(result)
						dict[key] = nil
						word = []byte{}
					} else {
						if key != "pieces" {
							value = string(result)
						} else {
							value = result
						}
					}
				}
			}
		}

		if len(key) > 0 && value != nil {
			dict[key] = value
			key = ""
			value = nil
			word = []byte{}
		}
	}
	return dict, bencode
}

func decode_list(bencode []byte) ([]any, error) {
	if bencode[0] != 'l' {
		return nil, fmt.Errorf("List must start with l")
	}
	result, _ := list_rec(bencode[1:])
	return result, nil
}

func list_rec(bencode []byte) ([]any, []byte) {
	list := []any{}
	word := []byte{}
	//iterate all letters
	for len(bencode) > 0 {
		char := bencode[0]
		bencode = bencode[1:] // -1 char

		if char == 'd' && len(word) == 0 { // if dict
			dict_res, code := dict_rec(bencode)
			bencode = code
			list = append(list, dict_res) // set value to list

		} else if char == 'e' && len(word) == 0 { //if end of list

			return list, bencode
		} else if char == 'l' && len(word) == 0 {
			dict_res, code := list_rec(bencode)
			bencode = code
			list = append(list, dict_res) // set value to list

		} else { // string or number
			word = append(word, char)
			if word[0] == 'i' && word[len(word)-1] == 'e' { // number
				number, _ := decode_int(string(word))
				list = append(list, number)
				word = []byte{}

			} else if unicode.IsDigit(rune(word[0])) {

				result, _ := decode_string(word)
				if result != nil {

					list = append(list, result)
					word = []byte{}
				}
			}
		}
	}
	return list, bencode

}
