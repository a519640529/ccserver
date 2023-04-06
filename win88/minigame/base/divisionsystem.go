package base

//提供税收和流水，根据代理需求后台进行分账
func ProfitDistribution(p *Player, tax, taxex, validFlow int64) {
	//LogChannelSington.WriteMQData(model.GenerateTaxDivide(p.SnId, p.Platform, p.Channel, p.BeUnderAgentCode, p.PackageID, tax, taxex, validFlow, p.scene.GameId, p.scene.GameMode, p.scene.DbGameFree.GetId(), p.PromoterTree))
}
