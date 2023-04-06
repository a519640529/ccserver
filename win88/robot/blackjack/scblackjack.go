package blackjack

import (
	rule "games.yol.com/win88/gamerule/blackjack"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/blackjack"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
	"reflect"
)

// 下注时间
func BlackJackBetTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{1, 3}
	case rule.CharacterB:
		return []int{3, 5}
	case rule.CharacterC:
		return []int{5, 7}
	default:
		return []int{3, 5}
	}
}

// 买保险时间
func BlackJackBuyTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{1, 3}
	case rule.CharacterB:
		return []int{3, 5}
	case rule.CharacterC:
		return []int{5, 7}
	default:
		return []int{3, 5}
	}
}

// 双倍时间
func BlackJackFenPaiTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{1, 3}
	case rule.CharacterB:
		return []int{3, 5}
	case rule.CharacterC:
		return []int{5, 7}
	default:
		return []int{3, 5}
	}
}

// 双倍时间
func BlackJackDoubleTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{1, 3}
	case rule.CharacterB:
		return []int{3, 5}
	case rule.CharacterC:
		return []int{5, 7}
	default:
		return []int{3, 5}
	}
}

// 补牌时间
func BlackJackOutsTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{3, 5}
	case rule.CharacterB:
		return []int{5, 7}
	case rule.CharacterC:
		return []int{7, 9}
	default:
		return []int{3, 5}
	}
}

// 过牌时间
func BlackJackSkipTime(n int) []int {
	switch n {
	case rule.CharacterA:
		return []int{1, 5}
	case rule.CharacterB:
		return []int{1, 5}
	case rule.CharacterC:
		return []int{1, 5}
	default:
		return []int{1, 5}
	}
}

// 下注类型
// 返回值 0 最小下注,1 最大下注, 2 随机下注
func BlackJackBetType(n int) int {
	num := rand.Intn(100)
	switch n {
	case rule.CharacterA:
		for _, v := range []int{20, 50, 100} {
			if num < v {
				switch v {
				case 20:
					return 0
				case 50:
					return 2
				case 100:
					return 1
				}
			}
		}
	case rule.CharacterB:
		for _, v := range []int{34, 67, 100} {
			if num < v {
				switch v {
				case 34:
					return 0
				case 67:
					return 2
				case 100:
					return 1
				}
			}
		}
	case rule.CharacterC:
		for _, v := range []int{50, 80, 100} {
			if num < v {
				switch v {
				case 50:
					return 0
				case 80:
					return 2
				case 100:
					return 1
				}
			}
		}
	}
	return 2
}

// 双倍概率
// n 性格
// x 点数
func BlackJackDoubleRate(n, x int) int {
	switch n {
	case rule.CharacterA:
		if 3 <= x && x <= 5 {
			return 20
		}
		if 6 <= x && x <= 8 {
			return 40
		}
		if 9 <= x && x <= 11 {
			return 60
		}
		if 12 <= x && x <= 14 {
			return 40
		}
		if 15 <= x && x <= 17 {
			return 20
		}
		if 18 == x {
			return 5
		}
		if 19 == x {
			return 1
		}
		if 20 == x {
			return 0
		}
	case rule.CharacterB:
		if 3 <= x && x <= 5 {
			return 0
		}
		if 6 <= x && x <= 8 {
			return 20
		}
		if 9 <= x && x <= 11 {
			return 80
		}
		if 12 <= x && x <= 14 {
			return 60
		}
		if 15 <= x && x <= 17 {
			return 20
		}
		if 18 <= x && x <= 20 {
			return 0
		}
	case rule.CharacterC:
		if 3 <= x && x <= 5 {
			return 0
		}
		if 6 <= x && x <= 8 {
			return 0
		}
		if 9 <= x && x <= 11 {
			return 60
		}
		if 12 <= x && x <= 14 {
			return 40
		}
		if 15 <= x && x <= 17 {
			return 0
		}
		if 18 <= x && x <= 20 {
			return 0
		}
	}
	return 0
}

// 要牌概率
// n 性格
// x 点数
func BlackJackOutsRate(n, x int) int {
	switch n {
	case rule.CharacterA:
		if 3 <= x && x <= 5 {
			return 100
		}
		if 6 <= x && x <= 8 {
			return 100
		}
		if 9 <= x && x <= 11 {
			return 100
		}
		if 12 <= x && x <= 14 {
			return 80
		}
		if 15 <= x && x <= 17 {
			return 60
		}
		if 18 == x {
			return 5
		}
		if 19 == x {
			return 1
		}
		if 20 == x {
			return 0
		}
	case rule.CharacterB:
		if 3 <= x && x <= 5 {
			return 100
		}
		if 6 <= x && x <= 8 {
			return 100
		}
		if 9 <= x && x <= 11 {
			return 100
		}
		if 12 <= x && x <= 14 {
			return 70
		}
		if 15 <= x && x <= 17 {
			return 50
		}
		if 18 <= x && x <= 20 {
			return 0
		}
	case rule.CharacterC:
		if 3 <= x && x <= 5 {
			return 100
		}
		if 6 <= x && x <= 8 {
			return 100
		}
		if 9 <= x && x <= 11 {
			return 100
		}
		if 12 <= x && x <= 14 {
			return 60
		}
		if 15 <= x && x <= 17 {
			return 40
		}
		if 18 <= x && x <= 20 {
			return 0
		}
	}
	return 0
}

func BlackJackRoomInfo(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackRoomInfo)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackRoomInfo error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackRoomInfo ", msg.String())
	scene, ok := base.SceneMgrSington.GetScene(msg.GetRoomId()).(*BlackJackScene)
	if !ok || scene == nil {
		scene = NewBlackJackScene(msg)
		base.SceneMgrSington.AddScene(scene)
	}
	if scene != nil {
		s.SetAttribute(base.SessionAttributeSceneId, scene.GetRoomId())
		scene.SCBlackJackRoomInfo = msg // 更新数据
		for _, pd := range msg.GetPlayers() {
			if pd.GetSnId() <= 0 {
				continue
			}
			if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
				if p := NewBlackJackPlayer(pd); p != nil {
					scene.AddPlayer(p)
				}
			} else {
				// 更新数据
				scene.players[pd.GetSnId()].BlackJackPlayer = pd
			}
		}
	}
	return nil
}

func BlackJackRoomState(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackRoomStatus)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackRoomState error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackRoomStatus ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	switch msg.GetStatus() {
	case rule.StatusBet:
		// 随机性格
		if p.character == 0 {
			p.character = rand.Intn(3) + 1
		}
		var betCoin int64
		// 随机下注类型（最小下注,最大下注,随机下注）
		switch BlackJackBetType(p.character) {
		case 0: // 最小下注
			betCoin = msg.GetParam()[0]
		case 1: // 最大下注
			if p.GetCoin() > msg.GetParam()[1] {
				betCoin = msg.GetParam()[1]
			} else {
				betCoin = p.GetCoin()
			}
		case 2: // 随机下注
			var top int64
			if p.GetCoin() > msg.GetParam()[1] {
				top = msg.GetParam()[1]
			} else {
				top = p.GetCoin()
			}
			betCoin = msg.GetParam()[0] + rand.Int63n(top-msg.GetParam()[0])
			betCoin = betCoin / 50 * 50 //因为最小下注额为1元（100分），下注单位只能精确到50分
		}
		if len(msg.GetParam()) > 2 { // 游服指定下注额
			betCoin = msg.GetParam()[2]
		}
		if betCoin > p.GetCoin() {
			betCoin = msg.GetParam()[0]
		}
		betCoin = (betCoin / 10) * 10
		p.betCoin = betCoin
		ret := blackjack.CSBlackJackOP{
			OpCode: proto.Int32(rule.SubBet),
			Params: []int64{int64(p.GetSeat()), betCoin},
		}
		base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), ret, BlackJackBetTime(p.character)...)
	}
	return nil
}

func BlackJackDeal(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackDeal)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackDeal error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackDeal ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	for _, c := range msg.GetSeats() {
		if p.GetSeat() == c.GetSeat() {
			p.Cards = []*blackjack.BlackJackCards{
				{
					Cards: c.GetCards(),
					Type:  c.Type,
					Point: c.GetPoint(),
					State: c.State,
					Id:    c.Id,
					Seat:  c.Seat,
				},
			}
			break
		}
	}
	return nil
}

func BlackJackBuy(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackNotifyBuy)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackNotifyBuy error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackNotifyBuy ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	if p.GetSeat() != msg.GetPos() {
		return nil
	}
	// 花费下注金额一半的费用
	var buy int64
	if p.betCoin+p.betCoin/2 <= p.GetCoin() {
		if rand.Intn(100) < 50 {
			buy = 1
			p.baoCoin = p.betCoin / 2
		}
	}
	ret := blackjack.CSBlackJackOP{
		OpCode: proto.Int32(rule.SubBuy),
		Params: []int64{buy},
	}
	base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), ret, BlackJackBuyTime(p.character)...)
	return nil
}

func BlackJackBuyEnd(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackEnd)
	if !ok {
		logger.Logger.Error("BlackJack SCBlackJackBuyEnd error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackBuyEnd ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	if p.baoCoin == 0 {
		return nil
	}
	for _, v := range msg.GetPlayers() {
		if v.GetPos() == p.GetSeat() {
			p.Coin = v.Coin
			p.baoCoin = 0
		}
	}

	return nil
}

func BlackJackNotifyOperatePro(s *netlib.Session, packetId int, data interface{}) error {
	return BlackJackNotifyOperate(s, packetId, data, false)
}

func BlackJackNotifyOperate(s *netlib.Session, packetId int, data interface{}, isOpAfter bool) error {
	msg, ok := data.(*blackjack.SCBlackJackNotifyOperate)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackNotifyOperate error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackNotifyOperate ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	if p.GetSeat() != msg.GetPos() {
		return nil
	}
	// 判断是否可以操作
	if p.GetCards()[0].GetState() == 1 && (len(p.GetCards()) <= 1 || p.GetCards()[1].GetState() == 1) {
		return nil
	}
	// 智能化运营
	if msg.GetCards() == "Smart" {
		pack := &blackjack.CSBlackJackOP{
			OpCode: proto.Int32(msg.GetSeat()),
		}
		var ts []int
		switch msg.GetSeat() {
		case rule.SubFenPai:
			ts = BlackJackFenPaiTime(p.character)
		case rule.SubDouble:
			ts = BlackJackDoubleTime(p.character)
		case rule.SubOuts:
			ts = BlackJackOutsTime(p.character)
		case rule.SubSkip:
			ts = BlackJackSkipTime(p.character)
		default:
			ts = BlackJackOutsTime(p.character)
		}
		base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), pack, ts...)
		return nil
	}
	// 分牌
	if len(p.GetCards()) == 1 &&
		len(p.GetCards()[0].GetCards()) == 2 &&
		p.GetCards()[0].GetCards()[0]%13 == p.GetCards()[0].GetCards()[1]%13 &&
		p.betCoin*2 <= p.GetCoin() { // 只有一手牌，并且没有停牌
		if rand.Intn(100) < 50 {
			pack := &blackjack.CSBlackJackOP{
				OpCode: proto.Int32(rule.SubFenPai),
			}
			base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), pack, BlackJackFenPaiTime(p.character)...)
			logger.Logger.Trace("BlackJack FenPai")
			return nil
		}
	}
	// 双倍
	for _, v := range p.GetCards() {
		if v.GetState() != 1 && len(v.GetCards()) == 2 && p.betCoin*2 <= p.GetCoin() {
			var flag bool
			for _, val := range v.GetPoint() {
				if val == 11 {
					flag = true
					break
				}
			}
			if flag && rand.Intn(100) < BlackJackDoubleRate(p.character, int(v.GetPoint()[len(v.GetPoint())-1])) {
				pack := &blackjack.CSBlackJackOP{
					OpCode: proto.Int32(rule.SubDouble),
				}
				base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), pack, BlackJackDoubleTime(p.character)...)
				logger.Logger.Trace("BlackJack Double")
				return nil
			}
		}
	}
	// 要牌
	for _, v := range p.GetCards() {
		length := len(p.GetCards())
		if v.GetState() != 1 && length < rule.MaxCardNum {
			if rand.Intn(100) < BlackJackOutsRate(p.character, int(v.GetPoint()[len(v.GetPoint())-1])) {
				pack := &blackjack.CSBlackJackOP{
					OpCode: proto.Int32(rule.SubOuts),
				}
				ts := BlackJackOutsTime(p.character)
				if (!isOpAfter && length == 1) || (isOpAfter && length == 2) {
					base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), pack, ts...)
				}
				return nil
			}
		}
	}
	pack := &blackjack.CSBlackJackOP{
		OpCode: proto.Int32(rule.SubSkip),
	}
	base.DelaySend(s, int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), pack, BlackJackSkipTime(p.character)...)
	logger.Logger.Trace("BlackJack Skip")
	return nil
}

// 同步机器人牌型数据，触发机器人操作
func BlackJackPlayerOperate(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackPlayerOperate)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackPlayerOperate error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackPlayerOperate ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	if p.GetSeat() != msg.GetPos() {
		return nil
	}
	if msg.GetCode() != 0 {
		return nil
	}
	switch msg.GetOperate() {
	case rule.SubFenPai, rule.SubDouble, rule.SubOuts: // 同步牌型数据
		for _, v := range msg.GetCards() {
			if len(p.Cards) <= int(v.GetId()) {
				p.Cards = append(p.Cards, new(blackjack.BlackJackCards))
			}
			hand := p.Cards[int(v.GetId())]
			hand.Cards = append(hand.Cards, v.GetCards()...)
			hand.DCards = proto.Int32(v.GetDCards())
			hand.Type = proto.Int32(v.GetType())
			hand.Point = v.GetPoint()
			hand.State = proto.Int32(v.GetState())
			hand.BetCoin = proto.Int64(v.GetBetCoin())
			if v.GetType() == rule.CardTypeBoom {
				p.Coin = msg.ReCoin
				p.betCoin += v.GetBetCoin()
			}
		}

	case rule.SubSkip, rule.SubSkipLeft, rule.SubSkipBomb:
		for _, v := range p.Cards {
			if v.GetState() != 1 {
				v.State = proto.Int32(1)
				break
			}
		}

	default:
		return nil
	}
	if msg.GetCardsStr() == "" {
		return nil
	}
	data = &blackjack.SCBlackJackNotifyOperate{
		Pos:   proto.Int32(msg.GetPos()),
		Cards: msg.CardsStr,
		Seat:  msg.Seat,
	}
	return BlackJackNotifyOperate(s, packetId, data, true)
}

func BlackJackEnd(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackEnd)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackEnd error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackEnd ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	player := scene.GetMe(s)
	p, ok := player.(*BlackJackPlayer)
	if !ok || p == nil {
		return nil
	}
	for _, v := range msg.GetPlayers() {
		if v.GetPos() == p.GetSeat() {
			p.Release()
			p.Coin = v.Coin
		}
	}
	return nil
}

func BlackJackPlayerEnter(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackPlayerEnter)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackPlayerEnter error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackPlayerEnter ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	if player := NewBlackJackPlayer(msg.GetPlayer()); player != nil {
		scene.AddPlayer(player)
	}
	return nil
}

func BlackJackPlayerLeave(s *netlib.Session, packetId int, data interface{}) error {
	msg, ok := data.(*blackjack.SCBlackJackPlayerLeave)
	if !ok {
		logger.Logger.Error("BlackJack BlackJackPlayerLeave error")
		return nil
	}
	logger.Logger.Trace("Recover BlackJack BlackJackPlayerLeave ", msg.String())
	scene, ok := base.GetScene(s).(*BlackJackScene)
	if !ok {
		return nil
	}
	if scene == nil {
		return nil
	}
	if player, ok := scene.GetPlayerByPos(msg.GetPos()[0]).(*BlackJackPlayer); ok && player != nil {
		scene.DelPlayer(player.GetSnId())
	}
	return nil
}

func init() {
	register := func(mainId int, msgType interface{}, h func(session *netlib.Session, packetId int, data interface{}) error) {
		f := func() interface{} {
			return reflect.New(reflect.TypeOf(msgType)).Interface()
		}
		netlib.RegisterFactory(mainId, netlib.PacketFactoryWrapper(f))
		netlib.RegisterHandler(mainId, netlib.HandlerWrapper(h))
	}
	// 创建房间
	register(int(blackjack.BlackJackPacketID_SC_ROOM_INFO), blackjack.SCBlackJackRoomInfo{}, BlackJackRoomInfo)
	// 状态改变:下注
	register(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), blackjack.SCBlackJackRoomStatus{}, BlackJackRoomState)
	// 发牌
	register(int(blackjack.BlackJackPacketID_SC_DEAL), blackjack.SCBlackJackDeal{}, BlackJackDeal)
	// 买保险
	register(int(blackjack.BlackJackPacketID_SC_NOTIFY_BUY), blackjack.SCBlackJackNotifyBuy{}, BlackJackBuy)
	// 保险结算
	register(int(blackjack.BlackJackPacketID_SC_BUY_END), blackjack.SCBlackJackEnd{}, BlackJackBuyEnd)
	// 闲家操作
	register(int(blackjack.BlackJackPacketID_SC_NOTIFY_OPERATE), blackjack.SCBlackJackNotifyOperate{}, BlackJackNotifyOperatePro)
	// 操作结果
	register(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), blackjack.SCBlackJackPlayerOperate{}, BlackJackPlayerOperate)
	// 结算
	register(int(blackjack.BlackJackPacketID_SC_END), blackjack.SCBlackJackEnd{}, BlackJackEnd)
	// 玩家进入
	register(int(blackjack.BlackJackPacketID_SC_PLAYER_ENTER), blackjack.SCBlackJackPlayerEnter{}, BlackJackPlayerEnter)
	// 玩家离开
	register(int(blackjack.BlackJackPacketID_SC_PLAYER_LEAVE), blackjack.SCBlackJackPlayerLeave{}, BlackJackPlayerLeave)
}
