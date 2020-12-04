package main

//type easy_type string
//
////json解析器的七种类型
//const (
//	EASY_NULL   easy_type = "EASY_NULL"
//	EASY_FALSE            = "EASY_FALSE"
//	EASY_TRUE             = "EASY_TRUE"
//	EASY_NUMBER           = "EASY_NUMBER"
//	EASY_STRING           = "EASY_STRING"
//	EASY_ARRAY            = "EASY_ARRAY"
//	EASY_OBJECT           = "EASY_OBJECT"
//)

//json解析器的七种类型
const (
	EASY_NULL = iota
	EASY_FALSE
	EASY_TRUE
	EASY_NUMBER
	EASY_STRING
	EASY_ARRAY
	EASY_OBJECT
)

//json树的节点
type Easy_value struct {
	vType int
}

//解析后的返回值
const (
	EASY_PARSE_OK = iota
	EASY_PARSE_EXPECT_VALUE
	EASY_PARSE_INVALID_VALUE
	EASY_PARSE_ROOT_NOT_SINGULAR
)
