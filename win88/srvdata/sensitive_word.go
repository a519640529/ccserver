package srvdata

var sensitiveWordsTree = make(map[rune]*Word)

type Word struct {
	w    []rune
	next map[rune]*Word
}

func initSensitiveWordTree() error {
	for _, word := range PBDB_Sensitive_WordsMgr.Datas.Arr {
		w := []rune(word.GetSensitive_Words())
		if len(w) == 0 {
			continue
		}
		AddSensitiveWord(w)
	}
	return nil
}

func AddSensitiveWord(w []rune) bool {
	if len(w) == 0 {
		return false
	}
	node := sensitiveWordsTree
	for i, c := range w {
		if _, exist := node[c]; !exist {
			if i == len(w)-1 {
				node[c] = &Word{w: w[i:], next: make(map[rune]*Word)}
			} else {
				node[c] = &Word{w: nil, next: make(map[rune]*Word)}
			}
		} else {
			if i == len(w)-1 {
				if _, exist := node[c]; exist {
					if node[c].w == nil {
						node[c].w = w[i:]
					}
					return true
				}
			}
		}
		node = node[c].next
	}
	return true
}

func DelSensitiveWord(w []rune) bool {
	if len(w) == 0 {
		return false
	}
	node := sensitiveWordsTree
	for i, c := range w {
		if _, exist := node[c]; !exist {
			return false
		}
		if i == len(w)-1 {
			if _, exist := node[c]; exist {
				node[c].w = nil
				return true
			}
		}
		node = node[c].next
	}
	return false
}

func HasSensitiveWord(words []rune) bool {
	cnt := len(words)
	for m := 0; m < cnt; m++ {
		node := sensitiveWordsTree
		for n := m; n < cnt; n++ {
			c := words[n]
			if content, exist := node[c]; exist {
				if content.w != nil {
					return true
				} else {
					node = content.next
				}
			} else {
				break
			}
		}
	}
	return false
}

func ReplaceSensitiveWord(words []rune) []rune {
	cnt := len(words)
	for m := 0; m < cnt; m++ {
		node := sensitiveWordsTree
		for n := m; n < cnt; n++ {
			c := words[n]
			if content, exist := node[c]; exist {
				if content.w != nil {
					for i := m; i <= n; i++ {
						words[i] = rune('*')
					}
				} else {
					node = content.next
				}
			} else {
				break
			}
		}
	}
	return words
}
