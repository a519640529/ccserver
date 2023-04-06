package base

type WaitOpPlayer struct {
	pos  int
	flag int
}

type WaitQuene struct {
	quene [8]WaitOpPlayer
	cnt   int
}

func (wq *WaitQuene) Compress() {
	if wq.cnt < 2 {
		return
	}
	var flag int
	//for i := 0; i < wq.cnt-1; {
	//	if wq.quene[i].pos == wq.quene[i+1].pos {
	//		flag = wq.quene[i].flag | wq.quene[i+1].flag
	//		copy(wq.quene[i:], wq.quene[i+1:wq.cnt])
	//		wq.quene[i].flag = flag
	//		wq.cnt--
	//		wq.quene[i+1].flag = 0
	//	} else {
	//		i++
	//	}
	//}
	for i := 0; i < wq.cnt; i++ {
		for j := i + 1; j < wq.cnt; j++ {
			if wq.quene[i].flag != 0 && wq.quene[i].pos == wq.quene[j].pos && wq.quene[j].flag != 0 {
				flag = wq.quene[i].flag | wq.quene[j].flag
				copy(wq.quene[i:], wq.quene[j:wq.cnt])
				wq.quene[i].flag = flag
				wq.cnt--
				wq.quene[j].flag = 0
			}
		}
	}

}
