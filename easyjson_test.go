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

func TestEasyParseNull(t *testing.T) {
	var v Easy_value
	v.vType = EASY_TRUE
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, "null"))
	expectEQInt(EASY_NULL, EasyGetType(&v))
}

func TestEasyParseTrue(t *testing.T) {
	var v Easy_value
	v.vType = EASY_TRUE
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, "true"))
	expectEQInt(EASY_TRUE, EasyGetType(&v))
}

func TestEasyParseFalse(t *testing.T) {
	var v Easy_value
	v.vType = EASY_TRUE
	expectEQInt(EASY_PARSE_OK, EasyParse(&v, "false"))
	expectEQInt(EASY_FALSE, EasyGetType(&v))

}

func TestParseExpectValue(t *testing.T) {
	var v Easy_value
	v.vType = EASY_FALSE
	expectEQInt(EASY_PARSE_EXPECT_VALUE, EasyParse(&v, ""))
	expectEQInt(EASY_NULL, EasyGetType(&v))

	v.vType = EASY_FALSE
	expectEQInt(EASY_PARSE_EXPECT_VALUE, EasyParse(&v, " "))
	expectEQInt(EASY_NULL, EasyGetType(&v))
}

func TestParseInvalidValue(t *testing.T) {
	var v Easy_value
	v.vType = EASY_FALSE
	expectEQInt(EASY_PARSE_INVALID_VALUE, EasyParse(&v, "nul"))
	expectEQInt(EASY_NULL, EasyGetType(&v))

	v.vType = EASY_FALSE
	expectEQInt(EASY_PARSE_INVALID_VALUE, EasyParse(&v, "?"))
	expectEQInt(EASY_NULL, EasyGetType(&v))
}

func TestParseRootNotSingular(t *testing.T) {
	var v Easy_value
	v.vType = EASY_FALSE
	expectEQInt(EASY_PARSE_ROOT_NOT_SINGULAR, EasyParse(&v, "null x"))
	expectEQInt(EASY_NULL, EasyGetType(&v))

}

func TestAll(t *testing.T) {
	log.Printf("%d/%d (%f%%) passed\n", testPass, testCount, float32(testPass*100.0/testCount))
}
