package dragonvstiger

import (
	"bytes"
	"encoding/gob"
)

type CardKindParam struct {
	flag int
}

func (this *CardKindParam) MarkFlag(flag int) {
	flag = 1 << uint(flag)
	this.flag |= flag
}
func (this *CardKindParam) UnmarkFlag(flag int) {
	flag = 1 << uint(flag)
	this.flag &= ^flag
}
func (this *CardKindParam) IsMarkFlag(flag int) bool {
	flag = 1 << uint(flag)
	if (this.flag & flag) != 0 {
		return true
	}
	return false
}
func (this *CardKindParam) GetFlag() int {
	return this.flag
}
func (this *CardKindParam) SetFlag(flag int) {
	this.flag = flag
}
func (this *CardKindParam) String() string {
	buff := ""
	for i := 0; i < CardsKind_Max; i++ {
		if this.IsMarkFlag(i) {
			buff += kindofcardstr[i]
			buff += "|"
		}
	}
	return buff
}
func (this *CardKindParam) Clone() *CardKindParam {
	ckp := &CardKindParam{}
	for i := CardsKind_Normal; i < CardsKind_Max; i++ {
		ckp.MarkFlag(i)
	}
	return ckp
}
func (this *CardKindParam) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(this)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (this *CardKindParam) Unmarshal(data []byte) error {
	md := &CardKindParam{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(md)
	if err != nil {
		return err
	} else {
		for i := CardsKind_Normal; i < CardsKind_Max; i++ {
			if md.IsMarkFlag(i) {
				this.MarkFlag(i)
			}
		}
		return nil
	}
}

type KindOfCard struct {
	kind     int
	maxValue int
	maxColor int
	cards    []int
}

func (this *KindOfCard) GetKind() int    { return this.kind }
func (this *KindOfCard) GetMax() int     { return this.maxValue }
func (this *KindOfCard) GetColor() int   { return this.maxColor }
func (this *KindOfCard) GetCards() []int { return this.cards }
func (this *KindOfCard) IsAAA() bool {
	return this.kind == CardsKind_ThreeSame || this.kind == CardsKind_Boom
}
func (this *KindOfCard) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(this)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (this *KindOfCard) Unmarshal(data []byte) error {
	koc := &KindOfCard{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(koc)
	if err != nil {
		return err
	} else {
		this.kind = koc.kind
		this.maxValue = koc.maxValue
		this.maxColor = koc.maxColor
		this.cards = koc.cards
		return nil
	}
}

var kindofcardstr = []string{
	"CardsKind_Normal",
	"CardsKind_Double",
	"CardsKind_ThreeSort",
	"CardsKind_SameColor",
	"CardsKind_A23",
	"CardsKind_SameColorSort",
	"CardsKind_ThreeSame",
	"CardsKind_235Double",
	"CardsKind_Boom",
	"CardsKind_235Boom",
	"CardsKind_Max",
}
