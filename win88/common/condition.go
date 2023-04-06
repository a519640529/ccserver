package common

import (
	"math"
	"strings"
)

const (
	CMP_OP_EQ  = 0
	CMP_OP_GT  = 1
	CMP_OP_LT  = 2
	CMP_OP_GTE = 3
	CMP_OP_LTE = 4
	CMP_OP_NEQ = 5
)

type Condition struct {
	ConditionID           int     //条件id
	ConditionCMP          int     //比较类型
	ConditionParamInt64   int64   //整形参数
	ConditionParamFloat64 float64 //浮点参数
	ConditionParamString  string  //字符串参数
}

func (this *Condition) CmpInt(n int64) bool {
	switch this.ConditionCMP {
	case CMP_OP_EQ:
		if n == this.ConditionParamInt64 {
			return true
		}
	case CMP_OP_GT:
		if n > this.ConditionParamInt64 {
			return true
		}
	case CMP_OP_LT:
		if n < this.ConditionParamInt64 {
			return true
		}
	case CMP_OP_GTE:
		if n >= this.ConditionParamInt64 {
			return true
		}
	case CMP_OP_LTE:
		if n <= this.ConditionParamInt64 {
			return true
		}
	case CMP_OP_NEQ:
		if n != this.ConditionParamInt64 {
			return true
		}
	}
	return false
}

func (this *Condition) CmpFloat64(n float64) bool {
	switch this.ConditionCMP {
	case CMP_OP_EQ:
		if n == this.ConditionParamFloat64 {
			return true
		}
	case CMP_OP_GT:
		if n > this.ConditionParamFloat64 {
			return true
		}
	case CMP_OP_LT:
		if n < this.ConditionParamFloat64 {
			return true
		}
	case CMP_OP_GTE:
		if n >= this.ConditionParamFloat64 {
			return true
		}
	case CMP_OP_LTE:
		if n <= this.ConditionParamFloat64 {
			return true
		}
	case CMP_OP_NEQ:
		if math.Abs(n-this.ConditionParamFloat64) > 0.00001 {
			return true
		}
	}
	return false
}

func (this *Condition) CmpString(n string) bool {
	switch this.ConditionCMP {
	case CMP_OP_EQ:
		if strings.Compare(n, this.ConditionParamString) == 0 {
			return true
		}
	case CMP_OP_NEQ:
		if strings.Compare(n, this.ConditionParamString) != 0 {
			return true
		}
	case CMP_OP_GT:
		if strings.Compare(n, this.ConditionParamString) > 0 {
			return true
		}
	case CMP_OP_LT:
		if strings.Compare(n, this.ConditionParamString) < 0 {
			return true
		}
	case CMP_OP_GTE:
		if strings.Compare(n, this.ConditionParamString) >= 0 {
			return true
		}
	case CMP_OP_LTE:
		if strings.Compare(n, this.ConditionParamString) <= 0 {
			return true
		}
	}
	return false
}

func (this *Condition) CmpBool(n bool) bool {
	cmpBool := false
	if this.ConditionParamInt64 > 0 {
		cmpBool = true
	}
	switch this.ConditionCMP {
	case CMP_OP_EQ:
		if n == cmpBool {
			return true
		}
	case CMP_OP_NEQ:
		if n != cmpBool {
			return true
		}
	}
	return false
}

//todo 检查参数,临时处理，暂时没想到更好的办法

var ConditionMgrSington = &ConditionMgr{
	pool: make(map[int]CheckBase),
}

type ConditionMgr struct {
	pool map[int]CheckBase
}

func (this *ConditionMgr) CheckGroup(need interface{}, con [][]*Condition) bool {
	if len(con) <= 0 {
		return false
	}
	//组内或，组间与
	for i := 0; i < len(con); i++ {
		cArray := con[i]
		isOk := false
		for j := 0; j < len(cArray); j++ {
			if this.check(need, cArray[j]) {
				isOk = true
				break
			}
		}
		if !isOk {
			return false
		}
	}

	return true
}

func (this *ConditionMgr) check(need interface{}, con *Condition) bool {
	c, ok := this.pool[con.ConditionID]
	if !ok {
		return false
	}

	return c.Check(need, con)
}

func (this *ConditionMgr) Register(cid int, c CheckBase) {
	this.pool[cid] = c
}

type CheckBase interface {
	Check(need interface{}, condition *Condition) bool
}

// ////////////////////////////////////////////////////////////
const (
	C_USER_PROMOTER    = 1 //判定用户推广员
	C_USER_ISNEW       = 2 //是否新用户
	C_USER_ISHAVETEL   = 3 //是否正式用户
	C_USER_ISFIRST     = 4 //是否首充
	C_USER_ISDAYFIRST  = 5 //是否当日首充
	C_USER_PAYEXCHANGE = 6 //充值提现比
	C_USER_PAYNNUM     = 7 //充值额度

)
