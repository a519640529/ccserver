package tienlen

import "games.yol.com/win88/api3th"

var Config = api3th.NewBaseConfig("TIENLEN", api3th.APITypeSmart)

const (
	Predict = "predict"
	Legal   = "legal"
)

func init() {
	Config.Register("PostFormTimeOut", Predict, "/predict")
	Config.Register("PostFormTimeOut", Legal, "/legal")
}

type PredictRequest struct {
	Bomb_num             int    `url:"bomb_num" form:"bomb_num"`
	Card_play_action_seq string `url:"card_play_action_seq" form:"card_play_action_seq"`
	Last_move_0          string `url:"last_move_0" form:"last_move_0"`
	Last_move_1          string `url:"last_move_1" form:"last_move_1"`
	Last_move_2          string `url:"last_move_2" form:"last_move_2"`
	Last_move_3          string `url:"last_move_3" form:"last_move_3"`
	Num_cards_left_0     int    `url:"num_cards_left_0" form:"num_cards_left_0"`
	Num_cards_left_1     int    `url:"num_cards_left_1" form:"num_cards_left_1"`
	Num_cards_left_2     int    `url:"num_cards_left_2" form:"num_cards_left_2"`
	Num_cards_left_3     int    `url:"num_cards_left_3" form:"num_cards_left_3"`
	Other_hand_cards     string `url:"other_hand_cards" form:"other_hand_cards"`
	Played_cards_0       string `url:"played_cards_0" form:"played_cards_0"`
	Played_cards_1       string `url:"played_cards_1" form:"played_cards_1"`
	Played_cards_2       string `url:"played_cards_2" form:"played_cards_2"`
	Played_cards_3       string `url:"played_cards_3" form:"played_cards_3"`
	Player_hand_cards    string `url:"player_hand_cards" form:"player_hand_cards"`
	Player_position      int    `url:"player_position" form:"player_position"`
}

type PredictResponse struct {
	Message  string            `json:"message"`
	Status   int               `json:"status"`
	Result   map[string]string `json:"result"`
	WinRates map[string]string `json:"win_rates"`
}

type LegalRequest struct {
	Player_hand_cards int    `url:"player_hand_cards" form:"player_hand_cards"`
	Rival_move        string `url:"rival_move" form:"rival_move"`
}

type LegalResponse struct {
	Legal_action string `json:"legal_action"`
	Message      string `json:"message"`
	Status       int    `json:"status"`
}

var CardToAiCard = map[int32]int32{
	-1:-1,0: 2, 1: 3, 2: 4, 3: 5, 4: 6, 5: 7, 6: 8, 7: 9, 8: 10, 9: 11, 10: 12, 11: 0, 12: 1,
	13: 15, 14: 16, 15: 17, 16: 18, 17: 19, 18: 20, 19: 21, 20: 22, 21: 23, 22: 24, 23: 25, 24: 13, 25: 14,
	26: 28, 27: 29, 28: 30, 29: 31, 30: 32, 31: 33, 32: 34, 33: 35, 34: 36, 35: 37, 36: 38, 37: 26, 38: 27,
	39: 41, 40: 42, 41: 43, 42: 44, 43: 45, 44: 46, 45: 47, 46: 48, 47: 49, 48: 50, 49: 51, 50: 39, 51: 40,
}
var AiCardToCard = map[int32]int32{
	-1:-1,2: 0, 3: 1, 4: 2, 5: 3, 6: 4, 7: 5, 8: 6, 9: 7, 10: 8, 11: 9, 12: 10, 0: 11, 1: 12,
	15: 13, 16: 14, 17: 15, 18: 16, 19: 17, 20: 18, 21: 19, 22: 20, 23: 21, 24: 22, 25: 23, 13: 24, 14: 25,
	28: 26, 29: 27, 30: 28, 31: 29, 32: 30, 33: 31, 34: 32, 35: 33, 36: 34, 37: 35, 38: 36, 26: 37, 27: 38,
	41: 39, 42: 40, 43: 41, 44: 42, 45: 43, 46: 44, 47: 45, 48: 46, 49: 47, 50: 48, 51: 49, 39: 50, 40: 51,
}
