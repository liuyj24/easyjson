package main

import (
	"errors"
	"strconv"
	"strings"
)

type easyContext struct {
	json      []byte
	stack     []byte
	size, top int
}

func EasyContextPushChar(c *easyContext, b byte) {
	c.stack = append(c.stack, b)
}

func EasyContextPush(c *easyContext, size int) int {
	ret := c.top
	c.top += size
	return ret
}

func EasyContextPop(c *easyContext) []byte {
	result := c.stack
	c.stack = c.stack[:0:0]
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
	switch c.json[0] {
	case 'n':
		return EasyParseLiteral(c, value, "null", EASY_NULL)
	case 't':
		return EasyParseLiteral(c, value, "true", EASY_TRUE)
	case 'f':
		return EasyParseLiteral(c, value, "false", EASY_FALSE)
	case '"':
		return EasyParseString(c, value)
	default:
		return EasyParseNum(c, value)
	}
}

func EasyParseString(c *easyContext, value *Easy_value) int {
	c.json = c.json[1:] //去掉第一个"
	i := 0
	for i < len(c.json) {
		switch c.json[i] {
		case '"':
			value.str = EasyContextPop(c)
			c.json = c.json[len(c.json):]
			value.vType = EASY_STRING
			return EASY_PARSE_OK
		case '\\':
			i++
			switch c.json[i] {
			case '"':
				EasyContextPushChar(c, '"')
			case '\\':
				EasyContextPushChar(c, '\\')
			case '/':
				EasyContextPushChar(c, '/')
			case 'b':
				EasyContextPushChar(c, '\b')
			case 'f':
				EasyContextPushChar(c, '\f')
			case 'n':
				EasyContextPushChar(c, '\n')
			case 'r':
				EasyContextPushChar(c, '\r')
			case 't':
				EasyContextPushChar(c, '\t')
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
			EasyContextPushChar(c, c.json[i])
		}
		i++
	}
	return EASY_PARSE_MISS_QUOTATION_MARK
}

func EasyParseUtf8(c *easyContext, u int64) {
	if u <= 0x7f {
		EasyContextPushChar(c, byte(u))
	} else if u <= 0x7FF {
		EasyContextPushChar(c, byte(0xc0|((u>>6)&0xff)))
		EasyContextPushChar(c, byte(0x80|(u&0x3f)))
	} else if u <= 0xFFFF {
		EasyContextPushChar(c, byte(0xe0|((u>>12)&0xff)))
		EasyContextPushChar(c, byte(0x80|((u>>6)&0x3f)))
		EasyContextPushChar(c, byte(0x80|(u&0x3f)))
	} else if u <= 0x10FFFF {
		EasyContextPushChar(c, byte(0xf0|((u>>18)&0xff)))
		EasyContextPushChar(c, byte(0x80|((u>>12)&0x3f)))
		EasyContextPushChar(c, byte(0x80|((u>>6)&0x3f)))
		EasyContextPushChar(c, byte(0x80|u&0x3f))
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
		if len(c.json) > 1 && !(c.json[1] == '.' || c.json[1] == 'e' || c.json[1] == 'E') {
			return EASY_PARSE_ROOT_NOT_SINGULAR
		}
	}

	if c.json[0] == '+' || c.json[0] == '.' || c.json[len(c.json)-1] == '.' || startWithLetter(c.json[0]) {
		return EASY_PARSE_INVALID_VALUE
	}

	f, err := strconv.ParseFloat(string(c.json), 64)
	if err != nil {
		//todo 错误判断处理是否有更好的方法
		if strings.Contains(err.Error(), strconv.ErrRange.Error()) {
			return EASY_PARSE_NUMBER_TOO_BIG
		}
		return EASY_PARSE_INVALID_VALUE
	}
	c.json = c.json[len(c.json):]
	value.num = f
	value.vType = EASY_NUMBER
	return EASY_PARSE_OK
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
