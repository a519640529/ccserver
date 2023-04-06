package base

const (
	CMP_LT  = 0 //小于
	CMP_LTE = 1 //小于等于
	CMP_EQ  = 2 //等于
	CMP_GT  = 3 //大于
	CMP_GTE = 4 //大于等于
)

func CmpInt32(value int32, cmpValue int32, cmp int) bool {
	switch cmp {
	case CMP_EQ:
		return value == cmpValue
	case CMP_LT:
		return value < cmpValue
	case CMP_LTE:
		return value <= cmpValue
	case CMP_GT:
		return value > cmpValue
	case CMP_GTE:
		return value >= cmpValue
	}

	return false
}
func CmpInt(value int, cmpValue int, cmp int) bool {
	switch cmp {
	case CMP_EQ:
		return value == cmpValue
	case CMP_LT:
		return value < cmpValue
	case CMP_LTE:
		return value <= cmpValue
	case CMP_GT:
		return value > cmpValue
	case CMP_GTE:
		return value >= cmpValue
	}

	return false
}
