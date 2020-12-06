package main

import (
	"errors"
	"strconv"
	"strings"
)

type easyContext struct {
	json      []byte
	bStack    []byte
	vStack    []EasyValue
	size, top int
}

func EasyParse(value *EasyValue, json string) int {
	var c easyContext
	c.json = []byte(json)
	value.vType = EASY_NULL
	easyParseWhitespace(&c)
	result := easyParseValue(&c, value)
	if result == EASY_PARSE_OK {
		easyParseWhitespace(&c)
		if len(c.json) > 0 {
			return EASY_PARSE_ROOT_NOT_SINGULAR
		}
	}
	return result
}

//去掉json前面的空格
func easyParseWhitespace(c *easyContext) {
	if len(c.json) == 0 {
		return
	}
	b := c.json[0]
	for b == ' ' || b == '\t' || b == '\n' || b == '\r' {
		c.json = c.json[1:]
		if len(c.json) == 0 {
			break
		}
		b = c.json[0]
	}
}

func easyParseValue(c *easyContext, value *EasyValue) int {
	if len(c.json) == 0 {
		return EASY_PARSE_EXPECT_VALUE
	}
	//easyParseWhitespace(c)
	switch c.json[0] {
	case 'n':
		return EasyParseLiteral(c, value, "null", EASY_NULL)
	case 't':
		return EasyParseLiteral(c, value, "true", EASY_TRUE)
	case 'f':
		return EasyParseLiteral(c, value, "false", EASY_FALSE)
	case '"':
		return EasyParseString(c, value)
	case '[':
		return EasyParseArray(c, value)
	case '{':
		return EasyParseObject(c, value)
	default:
		return EasyParseNum(c, value)
	}
}

func EasyParseObject(c *easyContext, value *EasyValue) int {
	var result int
	c.json = c.json[1:]
	easyParseWhitespace(c)

	//todo: 如果走进了这个if，意味着传入了一个空对象，EasyValue中的值是空的
	if c.json[0] == '}' {
		c.json = c.json[1:]
		value.vType = EASY_OBJECT
		return EASY_PARSE_OK
	}
	for {
		easyParseWhitespace(c)
		var obj EasyObj

		/* parse key */
		if len(c.json) < 1 || c.json[0] != '"' {
			result = EASY_PARSE_MISS_KEY
			break
		}
		//todo 把字符串保存起来
		if result, obj.key = EasyParseStringRaw(c); result != EASY_PARSE_OK {
			break
		}

		/* parse : */
		easyParseWhitespace(c)
		if c.json[0] != ':' {
			result = EASY_PARSE_MISS_COLON
			break
		}
		c.json = c.json[1:]

		/* parse value */
		easyParseWhitespace(c)
		if result = easyParseValue(c, &obj.value); result != EASY_PARSE_OK {
			break
		}
		value.o = append(value.o, obj)
		easyParseWhitespace(c)
		if len(c.json) < 1 {
			return EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET
		} else {
			if c.json[0] == ',' {
				c.json = c.json[1:]
			} else if c.json[0] == '}' {
				c.json = c.json[1:]
				value.vType = EASY_OBJECT
				return EASY_PARSE_OK
			} else {
				return EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET
			}
		}
	}
	return result
}

func EasyParseArray(c *easyContext, value *EasyValue) int {
	c.json = c.json[1:]
	easyParseWhitespace(c)

	//todo: 如果走进了这个if，意味着传入了一个空数组，EasyValue中的值是空的
	if c.json[0] == ']' {
		c.json = c.json[1:]
		value.vType = EASY_ARRAY
		return EASY_PARSE_OK
	}
	for {
		easyParseWhitespace(c)
		var v EasyValue
		if result := easyParseValue(c, &v); result != EASY_PARSE_OK {
			return result
		}

		value.e = append(value.e, v)
		easyParseWhitespace(c)

		if len(c.json) < 1 {
			return EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		} else {
			if c.json[0] == ',' {
				c.json = c.json[1:]
			} else if c.json[0] == ']' {
				c.json = c.json[1:]
				value.vType = EASY_ARRAY
				return EASY_PARSE_OK
			} else {
				return EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET
			}
		}
	}
}

func EasyParseStringRaw(c *easyContext) (int, string) {
	var result []byte
	c.json = c.json[1:] //去掉第一个"
	i := 0
	for i < len(c.json) {
		switch c.json[i] {
		case '"':
			c.json = c.json[i+1:]
			return EASY_PARSE_OK, string(result)
		case '\\':
			i++
			switch c.json[i] {
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			case '/':
				result = append(result, '/')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case 'u':
				ok, u := EasyParseHex4(c, i)
				if !ok {
					return EASY_PARSE_INVALID_UNICODE_HEX, ""
				}
				i += 4 //指针往后移动4格

				//判断是否要继续解析，因为有可能还有一个低代理项（也就是有连续两个/uxxxx/uxxxx）
				if u >= 0xD800 && u <= 0xDBFF {
					i++
					if c.json[i] != '\\' {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE, ""
					}
					i++
					if c.json[i] != 'u' {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE, ""
					}
					i++
					ok2, u2 := EasyParseHex4(c, i-1)
					if !ok2 {
						return EASY_PARSE_INVALID_UNICODE_HEX, ""
					}
					if u2 < 0xDC00 || u2 > 0xDFFF {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE, ""
					}
					u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000 //计算unicode码点
					i += 3
				}
				result = append(result, EasyParseUtf8(u)...)

			default:
				return EASY_PARSE_INVALID_STRING_ESCAPE, ""
			}
		default:
			if c.json[i] < 0x20 {
				return EASY_PARSE_INVALID_STRING_CHAR, ""
			}
			result = append(result, c.json[i])
		}
		i++
	}
	return EASY_PARSE_MISS_QUOTATION_MARK, ""
}

func EasyParseString(c *easyContext, value *EasyValue) int {
	result, str := EasyParseStringRaw(c)
	//其实不用判断也行
	if result != EASY_PARSE_OK {
		return result
	} else {
		value.vType = EASY_STRING
		value.str = []byte(str)
		return result
	}
}

func EasyParseUtf8(u int64) []byte {
	var tail []byte
	if u <= 0x7f {
		tail = append(tail, byte(u))
	} else if u <= 0x7FF {
		tail = append(tail, byte(0xc0|((u>>6)&0xff)))
		tail = append(tail, byte(0x80|(u&0x3f)))
	} else if u <= 0xFFFF {
		tail = append(tail, byte(0xe0|((u>>12)&0xff)))
		tail = append(tail, byte(0x80|((u>>6)&0x3f)))
		tail = append(tail, byte(0x80|(u&0x3f)))
	} else if u <= 0x10FFFF {
		tail = append(tail, byte(0xf0|((u>>18)&0xff)))
		tail = append(tail, byte(0x80|((u>>12)&0x3f)))
		tail = append(tail, byte(0x80|((u>>6)&0x3f)))
		tail = append(tail, byte(0x80|u&0x3f))
	}
	return tail
}

func EasyParseHex4(c *easyContext, index int) (bool, int64) {
	part := c.json[index+1 : index+5]
	i, err := strconv.ParseInt(string(part), 16, 64)
	if err != nil {
		return false, 0
	}
	return true, i
}

func EasyParseNum(c *easyContext, value *EasyValue) int {
	if c.json[0] == '0' {
		//下面这个if判断是用来过滤0后面只能跟.eE三种字符，否则不表示数字，但是在解释数组的时候这样的判断出现了麻烦
		//if len(c.json) > 1 && !(c.json[1] == '.' || c.json[1] == 'e' || c.json[1] == 'E') {
		//	return EASY_PARSE_ROOT_NOT_SINGULAR
		//}
		if len(c.json) > 1 && (c.json[1] == 'x' || (c.json[1] >= '0' && c.json[1] <= '9')) {
			return EASY_PARSE_ROOT_NOT_SINGULAR
		}
	}

	if c.json[0] == '+' || c.json[0] == '.' || c.json[len(c.json)-1] == '.' || startWithLetter(c.json[0]) {
		return EASY_PARSE_INVALID_VALUE
	}

	//下面这段出问题，用","来划分数字的结束不靠谱，还是一个个数吧
	////在把字节数组转为数字之前要判断后面是否有","，否则解析数组json串的时候会有问题
	//index := strings.Index(string(c.json), ",")
	//convStr := string(c.json)
	//if index >= 0 && index <= len(c.json) {
	//	convStr = string(c.json[0:index])
	//}
	//convStr = strings.TrimSpace(convStr)
	index := 0
	for index < len(c.json) {
		if isDigit(c.json[index]) {
			index++
		} else {
			break
		}
	}
	convStr := string(c.json[0:index])
	f, err := strconv.ParseFloat(convStr, 64)
	if err != nil {
		//todo 错误判断处理是否有更好的方法
		if strings.Contains(err.Error(), strconv.ErrRange.Error()) {
			return EASY_PARSE_NUMBER_TOO_BIG
		}
		return EASY_PARSE_INVALID_VALUE
	}
	c.json = c.json[len(convStr):]
	value.num = f
	value.vType = EASY_NUMBER
	return EASY_PARSE_OK
}

func isDigit(b byte) bool {
	if (b >= '0' && b <= '9') || b == 'e' || b == 'E' || b == '-' || b == '+' || b == '.' {
		return true
	} else {
		return false
	}
}

func startWithLetter(b byte) bool {
	if b >= '0' && b <= '9' || b == '-' {
		return false
	} else {
		return true
	}
}

func EasyParseLiteral(c *easyContext, value *EasyValue, target string, valueType int) int {
	ts := []byte(target)
	tsLen := len(ts)

	if len(c.json) < tsLen {
		return EASY_PARSE_INVALID_VALUE
	}
	for i := 0; i < tsLen; i++ {
		if c.json[i] != ts[i] {
			return EASY_PARSE_INVALID_VALUE
		}
	}
	c.json = c.json[tsLen:]
	value.vType = valueType
	return EASY_PARSE_OK
}

func EasyGetType(value *EasyValue) int {
	return value.vType
}

func EasyGetNum(value *EasyValue) (float64, error) {
	if value != nil && value.vType == EASY_NUMBER {
		return value.num, nil
	} else {
		return 0, errors.New("couldn't get the num")
	}
}

func EasySetString(value *EasyValue, str []byte) {
	value.str = str
	value.vType = EASY_STRING
}

func EasyFree(value *EasyValue) {
	value.vType = EASY_NULL
}
