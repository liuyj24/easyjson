package main

type easyContext struct {
	json []byte
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
		return EasyParseNull(c, value)
	case 't':
		return EasyParseTrue(c, value)
	case 'f':
		return EasyParseFalse(c, value)
	default:
		return EASY_PARSE_INVALID_VALUE
	}
}

func EasyParseNull(c *easyContext, value *Easy_value) int {
	if len(c.json) < 4 {
		return EASY_PARSE_INVALID_VALUE
	}
	if c.json[0] != 'n' || c.json[1] != 'u' || c.json[2] != 'l' || c.json[3] != 'l' {
		return EASY_PARSE_INVALID_VALUE
	}
	c.json = c.json[4:]
	value.vType = EASY_NULL

	if len(c.json) > 0 {
		return EASY_PARSE_ROOT_NOT_SINGULAR
	}
	return EASY_PARSE_OK
}

func EasyParseTrue(c *easyContext, value *Easy_value) int {
	if len(c.json) < 4 {
		return EASY_PARSE_INVALID_VALUE
	}
	if c.json[0] != 't' || c.json[1] != 'r' || c.json[2] != 'u' || c.json[3] != 'e' {
		return EASY_PARSE_INVALID_VALUE
	}
	c.json = c.json[4:]
	value.vType = EASY_TRUE
	if len(c.json) > 0 {
		return EASY_PARSE_ROOT_NOT_SINGULAR
	}
	return EASY_PARSE_OK
}

func EasyParseFalse(c *easyContext, value *Easy_value) int {
	if len(c.json) < 5 {
		return EASY_PARSE_INVALID_VALUE
	}
	if c.json[0] != 'f' || c.json[1] != 'a' || c.json[2] != 'l' || c.json[3] != 's' || c.json[4] != 'e' {
		return EASY_PARSE_INVALID_VALUE
	}
	c.json = c.json[5:]
	value.vType = EASY_FALSE
	if len(c.json) > 0 {
		return EASY_PARSE_ROOT_NOT_SINGULAR
	}
	return EASY_PARSE_OK

}

func EasyGetType(value *Easy_value) int {
	return value.vType
}
