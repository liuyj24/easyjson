package main

import (
	"errors"
	"strconv"
	"strings"
)

type easyContext struct {
	json      []byte
	bStack    []byte
	vStack    []Easy_value
	size, top int
}

func EasyContextBstackPush(c *easyContext, b byte) {
	c.bStack = append(c.bStack, b)
}

func EasyContextBstackPop(c *easyContext) []byte {
	result := c.bStack
	c.bStack = c.bStack[:0:0]
	return result
}

//func EasyContextPush(c *easyContext, size int) int {
//	ret := c.top
//	c.top += size
//	return ret
//}

func EasyContextVstackPush(c *easyContext, value Easy_value) {
	c.vStack = append(c.vStack, value)
}

func EasyContextVstackPop(c *easyContext) []Easy_value {
	result := c.vStack
	c.vStack = c.vStack[:0:0]
	return result
}

func EasyParse(value *Easy_value, json string) int {
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

func easyParseValue(c *easyContext, value *Easy_value) int {
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
	default:
		return EasyParseNum(c, value)
	}
}

func EasyParseArray(c *easyContext, value *Easy_value) int {
	c.json = c.json[1:]
	easyParseWhitespace(c)

	//TODO: 这里有可能出问题
	if c.json[0] == ']' {
		c.json = c.json[1:] //这里
		value.vType = EASY_ARRAY
		//value.e = nil //这里
		return EASY_PARSE_OK
	}
	for {
		easyParseWhitespace(c)
		var v Easy_value
		if result := easyParseValue(c, &v); result != EASY_PARSE_OK {
			return result
		}
		//解析一个值成功则放入栈中，如果是c语言的话，只用一个数组实现栈就可以了，但是其他好像不太行（我试试）
		//经过测试后确实不行，我觉得我不应该引入一个多余的栈
		//EasyContextVstackPush(c, v)

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
				//value.e = EasyContextVstackPop(c)
				return EASY_PARSE_OK
			} else {
				return EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET
			}
		}
	}
}

func EasyParseString(c *easyContext, value *Easy_value) int {
	c.json = c.json[1:] //去掉第一个"
	i := 0
	for i < len(c.json) {
		switch c.json[i] {
		case '"':
			value.str = EasyContextBstackPop(c)
			c.json = c.json[i+1:]
			value.vType = EASY_STRING
			return EASY_PARSE_OK
		case '\\':
			i++
			switch c.json[i] {
			case '"':
				EasyContextBstackPush(c, '"')
			case '\\':
				EasyContextBstackPush(c, '\\')
			case '/':
				EasyContextBstackPush(c, '/')
			case 'b':
				EasyContextBstackPush(c, '\b')
			case 'f':
				EasyContextBstackPush(c, '\f')
			case 'n':
				EasyContextBstackPush(c, '\n')
			case 'r':
				EasyContextBstackPush(c, '\r')
			case 't':
				EasyContextBstackPush(c, '\t')
			case 'u':
				ok, u := EasyParseHex4(c, i)
				if !ok {
					return EASY_PARSE_INVALID_UNICODE_HEX
				}
				i += 4 //指针往后移动4格

				//判断是否要继续解析，因为有可能还有一个低代理项（也就是有连续两个/uxxxx/uxxxx）
				if u >= 0xD800 && u <= 0xDBFF {
					i++
					if c.json[i] != '\\' {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE
					}
					i++
					if c.json[i] != 'u' {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE
					}
					i++
					ok2, u2 := EasyParseHex4(c, i-1)
					if !ok2 {
						return EASY_PARSE_INVALID_UNICODE_HEX
					}
					if u2 < 0xDC00 || u2 > 0xDFFF {
						return EASY_PARSE_INVALID_UNICODE_SURROGATE
					}
					u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000 //计算unicode码点

					//todo: 想想指针应该可以处理得更好
					i += 3
				}
				EasyParseUtf8(c, u)

			default:
				return EASY_PARSE_INVALID_STRING_ESCAPE
			}
		default:
			if c.json[i] < 0x20 {
				return EASY_PARSE_INVALID_STRING_CHAR
			}
			EasyContextBstackPush(c, c.json[i])
		}
		i++
	}
	return EASY_PARSE_MISS_QUOTATION_MARK
}

func EasyParseUtf8(c *easyContext, u int64) {
	if u <= 0x7f {
		EasyContextBstackPush(c, byte(u))
	} else if u <= 0x7FF {
		EasyContextBstackPush(c, byte(0xc0|((u>>6)&0xff)))
		EasyContextBstackPush(c, byte(0x80|(u&0x3f)))
	} else if u <= 0xFFFF {
		EasyContextBstackPush(c, byte(0xe0|((u>>12)&0xff)))
		EasyContextBstackPush(c, byte(0x80|((u>>6)&0x3f)))
		EasyContextBstackPush(c, byte(0x80|(u&0x3f)))
	} else if u <= 0x10FFFF {
		EasyContextBstackPush(c, byte(0xf0|((u>>18)&0xff)))
		EasyContextBstackPush(c, byte(0x80|((u>>12)&0x3f)))
		EasyContextBstackPush(c, byte(0x80|((u>>6)&0x3f)))
		EasyContextBstackPush(c, byte(0x80|u&0x3f))
	}
}

func EasyParseHex4(c *easyContext, index int) (bool, int64) {
	part := c.json[index+1 : index+5]
	i, err := strconv.ParseInt(string(part), 16, 64)
	if err != nil {
		return false, 0
	}
	return true, i
}

func EasyParseNum(c *easyContext, value *Easy_value) int {
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

func EasyParseLiteral(c *easyContext, value *Easy_value, target string, valueType int) int {
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

func EasyGetType(value *Easy_value) int {
	return value.vType
}

func EasyGetNum(value *Easy_value) (float64, error) {
	if value != nil && value.vType == EASY_NUMBER {
		return value.num, nil
	} else {
		return 0, errors.New("couldn't get the num")
	}
}

func EasySetString(value *Easy_value, str []byte) {
	value.str = str
	value.vType = EASY_STRING
}

func EasyFree(value *Easy_value) {
	value.vType = EASY_NULL
}
