package main

var _playerListeners []PlayerListener

func RegistePlayerListener(l PlayerListener) {
	for _, ll := range _playerListeners {
		if ll == l {
			return
		}
	}
	_playerListeners = append(_playerListeners, l)
}

type PlayerListener interface {
	//登录登出相关
	OnPlayerLogined(p *Player)
	OnPlayerLogouted(p *Player)
	OnPlayerDropLine(p *Player)
	OnPlayerRehold(p *Player)
	OnPlayerReturnScene(p *Player) //玩家返回房间
	//时间相关
	OnPlayerSecTimer(p *Player)
	OnPlayerMiniTimer(p *Player)
	OnPlayerHourTimer(p *Player)
	OnPlayerDayTimer(p *Player, login, continuous bool)
	OnPlayerWeekTimer(p *Player)
	OnPlayerMonthTimer(p *Player)
	//业务相关
	OnPlayerEnterScene(p *Player, s *Scene)
	OnPlayerLeaveScene(p *Player, s *Scene)
}

type BasePlayerListener struct {
}

func (l *BasePlayerListener) OnPlayerLogined(p *Player)                          {}
func (l *BasePlayerListener) OnPlayerLogouted(p *Player)                         {}
func (l *BasePlayerListener) OnPlayerDropLine(p *Player)                         {}
func (l *BasePlayerListener) OnPlayerRehold(p *Player)                           {}
func (l *BasePlayerListener) OnPlayerReturnScene(p *Player)                      {}
func (l *BasePlayerListener) OnPlayerSecTimer(p *Player)                         {}
func (l *BasePlayerListener) OnPlayerMiniTimer(p *Player)                        {}
func (l *BasePlayerListener) OnPlayerHourTimer(p *Player)                        {}
func (l *BasePlayerListener) OnPlayerDayTimer(p *Player, login, continuous bool) {}
func (l *BasePlayerListener) OnPlayerWeekTimer(p *Player)                        {}
func (l *BasePlayerListener) OnPlayerMonthTimer(p *Player)                       {}
func (l *BasePlayerListener) OnPlayerEnterScene(p *Player, s *Scene)             {}
func (l *BasePlayerListener) OnPlayerLeaveScene(p *Player, s *Scene)             {}

func FirePlayerLogined(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerLogined(p)
		}
	}
}

func FirePlayerLogouted(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerLogouted(p)
		}
	}
}

func FirePlayerDropLine(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerDropLine(p)
		}
	}
}

func FirePlayerRehold(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerRehold(p)
		}
	}
}

func FirePlayerReturnScene(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerReturnScene(p)
		}
	}
}

func FirePlayerSecTimer(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerSecTimer(p)
		}
	}
}

func FirePlayerMiniTimer(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerMiniTimer(p)
		}
	}
}

func FirePlayerHourTimer(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerHourTimer(p)
		}
	}
}

func FirePlayerDayTimer(p *Player, login, continuous bool) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerDayTimer(p, login, continuous)
		}
	}
}

func FirePlayerWeekTimer(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerWeekTimer(p)
		}
	}
}

func FirePlayerMonthTimer(p *Player) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerMonthTimer(p)
		}
	}
}

func FirePlayerEnterScene(p *Player, s *Scene) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerEnterScene(p, s)
		}
	}
}

func FirePlayerLeaveScene(p *Player, s *Scene) {
	for _, l := range _playerListeners {
		if l != nil {
			l.OnPlayerLeaveScene(p, s)
		}
	}
}
