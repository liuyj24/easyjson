package main

import (
	"log"
	"testing"
)

var mainRet = 0
var testCount = 0
var testPass = 0

func expectEQ(equality bool, expect interface{}, actual interface{}, format string) {
	testCount++
	if equality {
		testPass++
	} else {
		log.Printf("expect: "+format+", actual: "+format+"\n", expect, actual)
	}
}

func expectEQInt(expect int, actual int) {
	expectEQ(expect == actual, expect, actual, "%d")
}

func expectEQFloat(expect float64, actual float64) {
	expectEQ(expect == actual, expect, actual, "%f")
}

func expectEQString(expect string, actual string) {
	expectEQ(expect == actual, expect, actual, "%s")
}

//----- 测试字符串 -----

func TestParseString(t *testing.T) {
	testString("", "\"\"")
	testString("Hello", "\"Hello\"")
	testString("Hello\nWorld", "\"Hello\\nWorld\"")
	testString("\" \\ / \b \f \n \r \t", "\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"")
}

func TestParseMissingQuetationMark(t *testing.T) {
	ParseExpectValue(EASY_PARSE_MISS_QUOTATION_MARK, "\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_QUOTATION_MARK, "\"abc", EASY_NULL)
}

func TestParseInvalidStringEscape(t *testing.T) {
	ParseExpectValue(EASY_PARSE_INVALID_STRING_ESCAPE, "\"\\v\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_STRING_ESCAPE, "\"\\'\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_STRING_ESCAPE, "\"\\0\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_STRING_ESCAPE, "\"\\x12\"", EASY_NULL)
}

func TestParseInvalidStringChar(t *testing.T) {
	ParseExpectValue(EASY_PARSE_INVALID_STRING_CHAR, "\"\x01\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_STRING_CHAR, "\"\x1F\"", EASY_NULL)
}

func testString(expect string, json string) {
	var v Easy_value
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, json))
	expectEQInt(EASY_STRING, EasyGetType(&v))
	expectEQString(expect, string(v.str))
}

//----- 测试数字 -----

func TestParseNum(t *testing.T) {
	testNumber(0.0, "0")
	testNumber(0.0, "-0")
	testNumber(0.0, "-0.0")
	testNumber(1.0, "1")
	testNumber(-1.0, "-1")
	testNumber(1.5, "1.5")
	testNumber(-1.5, "-1.5")
	testNumber(3.1416, "3.1416")
	testNumber(1e10, "1E10")
	testNumber(1e10, "1e10")
	testNumber(1e+10, "1E+10")
	testNumber(1e-10, "1E-10")
	testNumber(-1e10, "-1E10")
	testNumber(-1e10, "-1e10")
	testNumber(-1e+10, "-1E+10")
	testNumber(-1e-10, "-1E-10")
	testNumber(1.234e+10, "1.234E+10")
	testNumber(1.234e-10, "1.234E-10")
	testNumber(0.0, "1e-10000") /* must underflow */
	testNumber(4.9406564584124654e-324, "4.9406564584124654E-324")
	testNumber(1.7976931348623157e-308, "1.7976931348623157e-308")
}

func TestParseNumToBig(t *testing.T) {
	ParseExpectValue(EASY_PARSE_NUMBER_TOO_BIG, "1e309", EASY_NULL)
	ParseExpectValue(EASY_PARSE_NUMBER_TOO_BIG, "-1e309", EASY_NULL)
}

func testNumber(num float64, json string) {
	var v Easy_value
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, json))
	expectEQInt(EASY_NUMBER, EasyGetType(&v))
	expectEQFloat(num, v.num)
}

//----- 测试null/true/false -----

func TestEasyParseNull(t *testing.T) {
	ParseExpectValue(EASY_PARSE_OK, "null", EASY_NULL)
}

func TestEasyParseTrue(t *testing.T) {
	ParseExpectValue(EASY_PARSE_OK, "true", EASY_TRUE)
}

func TestEasyParseFalse(t *testing.T) {
	ParseExpectValue(EASY_PARSE_OK, "false", EASY_FALSE)
}

func TestParseExpectValue(t *testing.T) {
	ParseExpectValue(EASY_PARSE_EXPECT_VALUE, "", EASY_NULL)
	ParseExpectValue(EASY_PARSE_EXPECT_VALUE, " ", EASY_NULL)
}

func TestParseInvalidValue(t *testing.T) {
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "nul", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "?", EASY_NULL)

	/* invalid number */
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "+0", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "+1", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, ".123", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "1.", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "INF", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "inf", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "NAN", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_VALUE, "nan", EASY_NULL) //fixed in judge if null
}

func TestParseRootNotSingular(t *testing.T) {
	ParseExpectValue(EASY_PARSE_ROOT_NOT_SINGULAR, "null x", EASY_NULL)
	ParseExpectValue(EASY_PARSE_ROOT_NOT_SINGULAR, "0123", EASY_NULL)
	ParseExpectValue(EASY_PARSE_ROOT_NOT_SINGULAR, "0x0", EASY_NULL)
	ParseExpectValue(EASY_PARSE_ROOT_NOT_SINGULAR, "0x123", EASY_NULL)
}

func ParseExpectValue(parseCode int, value string, valueType int) {
	var v Easy_value
	v.vType = EASY_FALSE
	expectEQInt(parseCode, EasyParse(&v, value))
	expectEQInt(valueType, EasyGetType(&v))
}

func TestAll(t *testing.T) {
	log.Printf("%d/%d (%f%%) passed, fail: %d\n", testPass, testCount, float32(testPass*100.0/testCount), testCount-testPass)
}
