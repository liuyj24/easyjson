package main

import (
	"log"
	"testing"
)

var mainRet = 0
var testCount = 0
var testPass = 0

//----- 测试对象 -----

func TestParseObject(t *testing.T) {
	var e EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&e, " { } "))
	expectEQInt(EASY_OBJECT, e.vType)

	var v EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&v,
		" { "+
			"\"n\" : null , "+
			"\"f\" : false , "+
			"\"t\" : true , "+
			"\"i\" : 123 , "+
			"\"s\" : \"abc\", "+
			"\"a\" : [ 1, 2, 3 ],"+
			"\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }"+
			" } "))
	expectEQInt(EASY_OBJECT, v.vType)
	expectEQInt(EASY_NULL, v.o[0].value.vType)
	expectEQInt(EASY_FALSE, v.o[1].value.vType)
	expectEQString("t", v.o[2].key)
	expectEQString("i", v.o[3].key)
	expectEQString("abc", string(v.o[4].value.str))
	expectEQInt(3, len(v.o[5].value.e))
	expectEQInt(3, len(v.o[6].value.o))
}

func TestMissKey(t *testing.T) {
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{1:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{true:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{false:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{null:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{[]:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{{}:1,", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_KEY, "{\"a\":1,", EASY_NULL)
}

func TestParseMissColon(t *testing.T) {
	ParseExpectValue(EASY_PARSE_MISS_COLON, "{\"a\"}", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COLON, "{\"a\",\"b\"}", EASY_NULL)
}

func TestParseMissCommaOrCurlyBracket(t *testing.T) {
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET, "{\"a\":1", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET, "{\"a\":1]", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET, "{\"a\":1 \"b\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_CURLY_BRACKET, "{\"a\":{}", EASY_NULL)
}

//----- 测试数组 -----

func TestParseArray(t *testing.T) {
	var v EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, "[ null , false , true , 123 , \"abc\" ]"))
	expectEQInt(EASY_ARRAY, v.vType)
	expectEQInt(EASY_NULL, v.e[0].vType)
	expectEQInt(EASY_FALSE, v.e[1].vType)
	expectEQInt(EASY_TRUE, v.e[2].vType)
	expectEQInt(EASY_NUMBER, v.e[3].vType)
	expectEQString("abc", string(v.e[4].str))

	var v2 EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&v2, "[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]"))
	expectEQInt(EASY_ARRAY, v2.e[0].vType)
	expectEQInt(3, len(v2.e[3].e))
	expectEQInt(2, int(v2.e[3].e[2].num))
}

func TestParseMissCommaOrSquareBracket(t *testing.T) {
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET, "[1", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET, "[1}", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET, "[1 2", EASY_NULL)
	ParseExpectValue(EASY_PARSE_MISS_COMMA_OR_SQUARE_BRACKET, "[[]", EASY_NULL)
}

//----- 测试字符串 -----

func TestParseString(t *testing.T) {
	testString("", "\"\"")
	testString("Hello", "\"Hello\"")
	testString("Hello\nWorld", "\"Hello\\nWorld\"")
	testString("\" \\ / \b \f \n \r \t", "\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"")
	testString("\x24", "\"\\u0024\"")
	testString("\xC2\xA2", "\"\\u00A2\"")
	testString("\xF0\x9D\x84\x9E", "\"\\uD834\\uDD1E\"")
	testString("\xF0\x9D\x84\x9E", "\"\\ud834\\udd1e\"")
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

func TestParseInvalidUnicodeHex(t *testing.T) {
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u0\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u01\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u012\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u/000\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\uG000\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u0/00\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u0G00\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u00/0\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u00G0\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u000/\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u000G\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_HEX, "\"\\u 123\"", EASY_NULL)
}

func TestParseInvalidUnicodeSurrogate(t *testing.T) {
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_SURROGATE, "\"\\uD800\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_SURROGATE, "\"\\uDBFF\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_SURROGATE, "\"\\uD800\\\\\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_SURROGATE, "\"\\uD800\\uDBFF\"", EASY_NULL)
	ParseExpectValue(EASY_PARSE_INVALID_UNICODE_SURROGATE, "\"\\uD800\\uE000\"", EASY_NULL)
}

func testString(expect string, json string) {
	var v EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, json))
	expectEQInt(EASY_STRING, EasyGetType(&v))
	expectEQString(expect, string(v.str))
}

//----- 测试数字 -----

func TestParseNum(t *testing.T) {
	testNumber(123, "123")
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
	var v EasyValue
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

/*----- 测试生成器 -----*/

func TestEasyStringifyValue(t *testing.T) {
	testRoundTrip("null")
	testRoundTrip("true")
	testRoundTrip("false")
}

func TestEasyStringifyNumber(t *testing.T) {
	testRoundTrip("0")
	testRoundTrip("-0")
	testRoundTrip("1")
	testRoundTrip("-1")
	testRoundTrip("1.5")
	testRoundTrip("-1.5")
	testRoundTrip("3.25")
	testRoundTrip("1e+20")
	testRoundTrip("1.234e+20")
	testRoundTrip("1.234e-20")
	testRoundTrip("1.0000000000000002")      /* the smallest number > 1 */
	testRoundTrip("4.9406564584124654e-324") /* minimum denormal */
	testRoundTrip("-4.9406564584124654e-324")
	testRoundTrip("2.2250738585072009e-308") /* Max subnormal double */
	testRoundTrip("-2.2250738585072009e-308")
	testRoundTrip("2.2250738585072014e-308") /* Min normal positive double */
	testRoundTrip("-2.2250738585072014e-308")
	testRoundTrip("1.7976931348623157e+308") /* Max double */
	testRoundTrip("-1.7976931348623157e+308")
}

func TestStringifyString(t *testing.T) {
	testRoundTrip("\"\"")
	testRoundTrip("\"Hello\"")
	testRoundTrip("\"Hello\\nWorld\"")
	testRoundTrip("\"\\\" \\\\ / \\b \\f \\n \\r \\t\"")
	testRoundTrip("\"Hello\\u0000World\"")
}

func TestStringifyArray(t *testing.T) {
	testRoundTrip("[]")
	testRoundTrip("[null,false,true,123,\"abc\",[1,2,3]]")
}

func TestStringifyObject(t *testing.T) {
	testRoundTrip("{}")
	testRoundTrip("{\"n\":null,\"f\":false,\"t\":true,\"i\":123,\"s\":\"abc\",\"a\":[1,2,3],\"o\":{\"1\":1,\"2\":2,\"3\":3}}")

}

func testRoundTrip(json string) {
	var v EasyValue
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, json))
	json2 := EasyStringify(&v)
	expectEQString(json, json2)
}

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

func ParseExpectValue(parseCode int, value string, valueType int) {
	var v EasyValue
	v.vType = EASY_FALSE
	expectEQInt(parseCode, EasyParse(&v, value))
	expectEQInt(valueType, EasyGetType(&v))
}

func TestInfo(t *testing.T) {
	log.Printf("%d/%d (%f%%) passed, fail: %d\n",
		testPass, testCount, float32(testPass*100.0/testCount), testCount-testPass)
}
