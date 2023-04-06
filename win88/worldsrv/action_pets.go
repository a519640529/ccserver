package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/protocol/pets"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSRoleInfoPacketFactory struct {
}

type CSRoleInfoHandler struct {
}

func (this *CSRoleInfoPacketFactory) CreatePacket() interface{} {
	pack := &pets.CSRoleInfo{}
	return pack
}

func (this *CSRoleInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRoleInfoHandler Process recv ", data)
	if _, ok := data.(*pets.CSRoleInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRoleInfoHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		roleInfos := PetMgrSington.GetRoleInfos(p)
		pack := &pets.SCRoleInfo{
			Infos: roleInfos,
		}
		logger.Logger.Trace("SCRoleInfo:", pack)
		p.SendToClient(int(pets.PetsPacketID_PACKET_SC_ROLE_INFO), pack)
	}

	return nil
}

type CSPetInfoPacketFactory struct {
}

type CSPetInfoHandler struct {
}

func (this *CSPetInfoPacketFactory) CreatePacket() interface{} {
	pack := &pets.CSPetInfo{}
	return pack
}

func (this *CSPetInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPetInfoHandler Process recv ", data)
	if _, ok := data.(*pets.CSPetInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPetInfoHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}

		petInfos := PetMgrSington.GetPetInfos(p)
		pack := &pets.SCPetInfo{
			Infos: petInfos,
		}
		logger.Logger.Trace("SCPetInfo:", pack)
		p.SendToClient(int(pets.PetsPacketID_PACKET_SC_PET_INFO), pack)
	}
	return nil
}

type CSRisingStarPacketFactory struct {
}

type CSRisingStarHandler struct {
}

func (this *CSRisingStarPacketFactory) CreatePacket() interface{} {
	pack := &pets.CSRisingStar{}
	return pack
}

func (this *CSRisingStarHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRisingStarHandler Process recv ", data)
	if msg, ok := data.(*pets.CSRisingStar); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRisingStarHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		SendInfoRole := func(retCode pets.OpResultCode, roleInfo *pets.RoleInfo) {
			pack := &pets.SCRoleRisingStar{
				RetCode:  retCode,
				RoleInfo: roleInfo,
			}
			logger.Logger.Trace("SCPetRisingStar:", pack)
			p.SendToClient(int(pets.PetsPacketID_PACKET_SC_ROLE_RISINGSTAR), pack)
		}
		SendInfoPet := func(retCode pets.OpResultCode, petInfo *pets.PetInfo) {
			pack := &pets.SCPetRisingStar{
				RetCode: retCode,
				PetInfo: petInfo,
			}
			logger.Logger.Trace("SCPetRisingStar:", pack)
			p.SendToClient(int(pets.PetsPacketID_PACKET_SC_PET_RISINGSTAR), pack)
		}
		if msg.RisingModId == 0 {
			logger.Logger.Warn("CSRisingStarHandler UseModId:", msg.RisingModId)
			SendInfoRole(pets.OpResultCode_OPRC_Error, PetMgrSington.GetRoleInfo(p, msg.RisingModId))
			return nil
		}
		if msg.RisingType == 0 {
			roleInfo := PetMgrSington.GetIntroductionByModId(msg.RisingModId)
			if roleInfo == nil {
				SendInfoRole(pets.OpResultCode_OPRC_Error, PetMgrSington.GetRoleInfo(p, msg.RisingModId))
				return nil
			}
			if roleInfo.MaxLevel == p.Roles.ModUnlock[msg.RisingModId] {
				logger.Logger.Trace("人物已经达到最大等级")
				SendInfoRole(pets.OpResultCode_OPRC_Error, PetMgrSington.GetRoleInfo(p, msg.RisingModId))
				return nil
			}
			role := PetMgrSington.GetRoleInfo(p, msg.RisingModId)
			if role != nil {
				if role.HaveAmount < role.Amount {
					logger.Logger.Trace("人物碎片道具数量不够", role.HaveAmount, role.Amount)
					return nil
				}
			}
			//背包数据处理
			item := BagMgrSington.GetBagItemById(p.SnId, role.Fragment)
			if item != nil {
				item.ItemNum -= role.Amount
				role.HaveAmount -= role.Amount
			}
			//人物模型状态处理
			p.Roles.ModUnlock[msg.RisingModId]++
			FriendMgrSington.UpdateFriendRoles(p.SnId, p.Roles.ModUnlock)
			p.dirty = true
			//人物
			SendInfoRole(pets.OpResultCode_OPRC_Sucess, PetMgrSington.GetRoleInfo(p, msg.RisingModId))
			remark := role.Name + "升星"
			BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, role.Amount, remark)

			BagMgrSington.SyncBagData(p, item.ItemId)
		} else if msg.RisingType == 1 {
			petInfo := PetMgrSington.GetIntroductionByModId(msg.RisingModId)
			if petInfo == nil {
				SendInfoPet(pets.OpResultCode_OPRC_Error, PetMgrSington.GetPetInfo(p, msg.RisingModId))
				return nil
			}
			if petInfo.MaxLevel == p.Pets.ModUnlock[msg.RisingModId] {
				logger.Logger.Trace("宠物已经达到最大等级")
				SendInfoPet(pets.OpResultCode_OPRC_Error, PetMgrSington.GetPetInfo(p, msg.RisingModId))
				return nil
			}

			pet := PetMgrSington.GetPetInfo(p, msg.RisingModId)
			if pet != nil {
				if pet.HaveAmount < pet.Amount {
					logger.Logger.Trace("宠物碎片道具数量不够", pet.HaveAmount, pet.Amount)
					return nil
				}
			}
			//背包数据处理
			item := BagMgrSington.GetBagItemById(p.SnId, pet.Fragment)
			if item != nil {
				item.ItemNum -= pet.Amount
				pet.HaveAmount -= pet.Amount
			}

			p.Pets.ModUnlock[msg.RisingModId]++
			FriendMgrSington.UpdateFriendPets(p.SnId, p.Pets.ModUnlock)
			p.dirty = true
			//宠物
			SendInfoPet(pets.OpResultCode_OPRC_Sucess, PetMgrSington.GetPetInfo(p, msg.RisingModId))
			remark := pet.Name + "升星"
			BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, pet.Amount, remark)

			BagMgrSington.SyncBagData(p, item.ItemId)
		}

	}
	return nil
}

type CSRolePetUseOpPacketFactory struct {
}

type CSRolePetUseOpHandler struct {
}

func (this *CSRolePetUseOpPacketFactory) CreatePacket() interface{} {
	pack := &pets.CSRolePetUseOp{}
	return pack
}

func (this *CSRolePetUseOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRolePetUseOpHandler Process recv ", data)
	if msg, ok := data.(*pets.CSRolePetUseOp); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRolePetUseOpHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		if msg.UseModId == 0 {
			logger.Logger.Warn("CSRolePetUseOpHandler UseModId:", msg.UseModId)
			return nil
		}
		if msg.UseModType == 0 {
			if p.Roles.ModId == msg.UseModId {
				logger.Logger.Trace("人物使用中 不能直接取消人物使用")
				return nil
			}
			p.Roles.ModId = msg.UseModId
			p.dirty = true
			logger.Logger.Trace("使用人物:", msg.UseModId)
		} else {
			if p.Pets.ModId == msg.UseModId {
				p.Pets.ModId = 0
				logger.Logger.Trace("取消宠物跟随:", msg.UseModId)
				p.dirty = true
			} else {
				logger.Logger.Trace("设置宠物跟随:", msg.UseModId)
				p.Pets.ModId = msg.UseModId
				p.dirty = true
			}
		}
		pack := &pets.SCRolePetUseOp{
			RetCode:    pets.OpResultCode_OPRC_Sucess,
			UseModType: msg.UseModType,
			UseModId:   msg.UseModId,
		}
		logger.Logger.Trace("SCRolePetUseOp:", pack)
		p.SendToClient(int(pets.PetsPacketID_PACKET_SC_ROLEPETUSEOP), pack)
	}
	return nil
}

type CSRolePetUnlockPacketFactory struct {
}

type CSRolePetUnlockHandler struct {
}

func (this *CSRolePetUnlockPacketFactory) CreatePacket() interface{} {
	pack := &pets.CSRolePetUnlock{}
	return pack
}

func (this *CSRolePetUnlockHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRolePetUnlockHandler Process recv ", data)
	if msg, ok := data.(*pets.CSRolePetUnlock); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRolePetUnlockHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		SendMsg := func(retCode pets.OpResultCode, roleInfo *pets.RoleInfo, petInfo *pets.PetInfo) {
			pack := &pets.SCRolePetUnlock{
				RetCode:    retCode,
				UseModType: msg.UseModType,
				UseModId:   msg.UseModId,
				RoleInfo:   roleInfo,
				PetInfo:    petInfo,
			}
			logger.Logger.Trace("SCRolePetUnlock:", pack)
			p.SendToClient(int(pets.PetsPacketID_PACKET_SC_ROLEPETUNLOCK), pack)
		}
		if msg.UseModId == 0 {
			logger.Logger.Warn("CSRolePetUnlockHandler UseModId:", msg.UseModId)
			SendMsg(pets.OpResultCode_OPRC_Error, nil, nil)
			return nil
		}
		if msg.UseModType == 0 {
			if _, ok1 := p.Roles.ModUnlock[msg.UseModId]; !ok1 {
				roleInfo := PetMgrSington.GetRoleInfo(p, msg.UseModId)
				if roleInfo != nil {
					item := BagMgrSington.GetBagItemById(p.SnId, roleInfo.Fragment)
					if item != nil && item.ItemNum >= roleInfo.Amount {
						item.ItemNum -= roleInfo.Amount
						p.Roles.ModUnlock[msg.UseModId] = 1
						FriendMgrSington.UpdateFriendRoles(p.SnId, p.Roles.ModUnlock)
						p.dirty = true
						logger.Logger.Trace("解锁人物", msg.UseModId)
						SendMsg(pets.OpResultCode_OPRC_Sucess, PetMgrSington.GetRoleInfo(p, msg.UseModId), nil)
						remark := roleInfo.Name + "解锁"
						BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, roleInfo.Amount, remark)
						return nil
					}
				}
			}
		} else if msg.UseModType == 1 {
			if _, ok1 := p.Pets.ModUnlock[msg.UseModId]; !ok1 {
				petInfo := PetMgrSington.GetPetInfo(p, msg.UseModId)
				if petInfo != nil {
					item := BagMgrSington.GetBagItemById(p.SnId, petInfo.Fragment)
					if item != nil && item.ItemNum >= petInfo.Amount {
						item.ItemNum -= petInfo.Amount
						p.Pets.ModUnlock[msg.UseModId] = 1
						FriendMgrSington.UpdateFriendPets(p.SnId, p.Pets.ModUnlock)
						p.dirty = true
						logger.Logger.Trace("解锁宠物", msg.UseModId)
						SendMsg(pets.OpResultCode_OPRC_Sucess, nil, PetMgrSington.GetPetInfo(p, msg.UseModId))
						remark := petInfo.Name + "解锁"
						BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, petInfo.Amount, remark)
						return nil
					}
				}
			}
		}
		SendMsg(pets.OpResultCode_OPRC_Error, nil, nil)
	}
	return nil
}
func init() {
	common.RegisterHandler(int(pets.PetsPacketID_PACKET_CS_ROLE_INFO), &CSRoleInfoHandler{})
	netlib.RegisterFactory(int(pets.PetsPacketID_PACKET_CS_ROLE_INFO), &CSRoleInfoPacketFactory{})
	common.RegisterHandler(int(pets.PetsPacketID_PACKET_CS_PET_INFO), &CSPetInfoHandler{})
	netlib.RegisterFactory(int(pets.PetsPacketID_PACKET_CS_PET_INFO), &CSPetInfoPacketFactory{})
	common.RegisterHandler(int(pets.PetsPacketID_PACKET_CS_PET_RISINGSTAR), &CSRisingStarHandler{})
	netlib.RegisterFactory(int(pets.PetsPacketID_PACKET_CS_PET_RISINGSTAR), &CSRisingStarPacketFactory{})
	common.RegisterHandler(int(pets.PetsPacketID_PACKET_CS_ROLEPETUSEOP), &CSRolePetUseOpHandler{})
	netlib.RegisterFactory(int(pets.PetsPacketID_PACKET_CS_ROLEPETUSEOP), &CSRolePetUseOpPacketFactory{})
	common.RegisterHandler(int(pets.PetsPacketID_PACKET_CS_ROLEPETUNLOCK), &CSRolePetUnlockHandler{})
	netlib.RegisterFactory(int(pets.PetsPacketID_PACKET_CS_ROLEPETUNLOCK), &CSRolePetUnlockPacketFactory{})
}
