package main

import (
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"time"

	"games.yol.com/win88/protocol/pets"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/module"
)

const (
	ItemObtain  = iota //得到
	ItemConsume        //消耗
)

var PetMgrSington = &PetMgr{
	//RPInfos: make(map[int32]*RoleAndPetInfo),
}

type PetMgr struct {
	//RPInfos map[int32]*RoleAndPetInfo
}
type RoleAndPetInfo struct {
	ModId      int32  //模型id
	Name       string //名字
	Story      string //人物背景介绍
	AwardTitle string //奖励标题
	MaxLevel   int32  //最大等级
}

func (this *PetMgr) ModuleName() string {
	return "PetMgr"
}

// 人物
func (this *PetMgr) GetRoleInfos(p *Player) []*pets.RoleInfo {
	p.InitRolesAndPets()

	rolesInfo := srvdata.PBDB_Game_RoleMgr.Datas.Arr
	if rolesInfo != nil {
		var newRoles = make([]*pets.RoleInfo, 0)
		roleMaps := make(map[int32]*pets.RoleInfo)
		for k, roleInfo := range rolesInfo {
			role := &pets.RoleInfo{
				Id:        roleInfo.Id,
				RoleId:    roleInfo.RoleId,
				Name:      roleInfo.Name,
				Grade:     roleInfo.Grade,
				Level:     roleInfo.Level,
				Fragment:  roleInfo.Fragment,
				Amount:    roleInfo.Amount,
				AwardType: roleInfo.AwardType,
				Award:     roleInfo.Award,
				AwardRate: roleInfo.AwardRate,
			}
			if _, ok := p.Roles.ModUnlock[roleInfo.RoleId]; ok {
				role.IsUnlock = true
				if p.Roles.ModId == roleInfo.RoleId {
					role.IsUsing = true
				}
			}
			level := p.Roles.ModUnlock[roleInfo.RoleId]
			if role.Level == level {
				var nextAward int32
				if k+1 < len(rolesInfo) {
					nextAward = rolesInfo[k+1].Award
				}
				role.NextAward = nextAward
				roleMaps[roleInfo.RoleId] = role
			}
		}
		for _, vRole := range roleMaps {
			roleOther := this.GetIntroductionByModId(vRole.RoleId)
			vRole.MaxLevel = roleOther.MaxLevel
			vRole.Name = roleOther.Name
			vRole.Story = roleOther.Story
			vRole.AwardTitle = roleOther.AwardTitle
			///
			item := BagMgrSington.GetBagItemById(p.SnId, vRole.Fragment)
			if item != nil {
				vRole.HaveAmount = item.ItemNum
			}
			newRoles = append(newRoles, vRole)
		}

		return newRoles
	}
	return nil
}
func (this *PetMgr) GetRoleInfo(p *Player, modId int32) *pets.RoleInfo {
	p.InitRolesAndPets()
	rolesInfo := srvdata.PBDB_Game_RoleMgr.Datas.GetArr()
	if rolesInfo != nil {
		var newRole *pets.RoleInfo
		for k, roleInfo := range rolesInfo {
			if roleInfo.RoleId != modId {
				continue
			}
			role := &pets.RoleInfo{
				Id:        roleInfo.Id,
				RoleId:    roleInfo.RoleId,
				Name:      roleInfo.Name,
				Grade:     roleInfo.Grade,
				Level:     roleInfo.Level,
				Fragment:  roleInfo.Fragment,
				Amount:    roleInfo.Amount,
				AwardType: roleInfo.AwardType,
				Award:     roleInfo.Award,
				AwardRate: roleInfo.AwardRate,
			}
			if _, ok := p.Roles.ModUnlock[roleInfo.RoleId]; ok {
				role.IsUnlock = true
				if p.Roles.ModId == roleInfo.RoleId {
					role.IsUsing = true
				}
			}
			level := p.Roles.ModUnlock[roleInfo.RoleId]
			if role.Level == level && role.RoleId == modId {
				var nextAward int32
				if k+1 < len(rolesInfo) {
					nextAward = rolesInfo[k+1].Award
				}
				role.NextAward = nextAward
				newRole = role
				break
			}
		}
		roleOther := this.GetIntroductionByModId(modId)
		newRole.MaxLevel = roleOther.MaxLevel
		newRole.Name = roleOther.Name
		newRole.Story = roleOther.Story
		newRole.AwardTitle = roleOther.AwardTitle
		//
		item := BagMgrSington.GetBagItemById(p.SnId, newRole.Fragment)
		if item != nil {
			newRole.HaveAmount = item.ItemNum
		}
		return newRole
	}
	return nil
}

// ////////////////////////////////////////////////////////////////////////
// 宠物
func (this *PetMgr) GetPetInfos(p *Player) []*pets.PetInfo {
	p.InitRolesAndPets()

	petsInfo := srvdata.PBDB_Game_PetMgr.Datas.Arr
	if petsInfo != nil {
		var newPets = make([]*pets.PetInfo, 0)
		petMaps := make(map[int32]*pets.PetInfo)
		for k, petInfo := range petsInfo {
			pet := &pets.PetInfo{
				Id:        petInfo.Id,
				PetId:     petInfo.PetId,
				Name:      petInfo.Name,
				Grade:     petInfo.Grade,
				Level:     petInfo.Level,
				Fragment:  petInfo.Fragment,
				Amount:    petInfo.Amount,
				AwardType: petInfo.AwardType,
				Award:     petInfo.Award,
				AwardRate: petInfo.AwardRate,
			}
			if _, ok := p.Pets.ModUnlock[petInfo.PetId]; ok {
				pet.IsUnlock = true
				if p.Pets.ModId == petInfo.PetId {
					pet.IsUsing = true
				}
			}
			level := p.Pets.ModUnlock[petInfo.PetId]
			if pet.Level == level {
				var nextAward int32
				if k+1 < len(petsInfo) {
					nextAward = petsInfo[k+1].Award
				}
				pet.NextAward = nextAward
				petMaps[petInfo.PetId] = pet
			}
		}
		for _, vPet := range petMaps {
			petOther := this.GetIntroductionByModId(vPet.PetId)
			vPet.MaxLevel = petOther.MaxLevel
			vPet.Name = petOther.Name
			vPet.Story = petOther.Story
			vPet.AwardTitle = petOther.AwardTitle
			//
			item := BagMgrSington.GetBagItemById(p.SnId, vPet.Fragment)
			if item != nil {
				vPet.HaveAmount = item.ItemNum
			}
			newPets = append(newPets, vPet)
		}
		return newPets
	}
	return nil
}
func (this *PetMgr) GetPetInfo(p *Player, modId int32) *pets.PetInfo {
	p.InitRolesAndPets()
	petsInfo := srvdata.PBDB_Game_PetMgr.Datas.GetArr()
	if petsInfo != nil {
		var newPets *pets.PetInfo
		for k, petInfo := range petsInfo {
			if petInfo.PetId != modId {
				continue
			}
			pet := &pets.PetInfo{
				Id:        petInfo.Id,
				PetId:     petInfo.PetId,
				Name:      petInfo.Name,
				Grade:     petInfo.Grade,
				Level:     petInfo.Level,
				Fragment:  petInfo.Fragment,
				Amount:    petInfo.Amount,
				AwardType: petInfo.AwardType,
				Award:     petInfo.Award,
				AwardRate: petInfo.AwardRate,
			}
			if _, ok := p.Pets.ModUnlock[petInfo.PetId]; ok {
				pet.IsUnlock = true
				if p.Pets.ModId == petInfo.PetId {
					pet.IsUsing = true
				}
			}
			level := p.Pets.ModUnlock[petInfo.PetId]
			if pet.Level == level && pet.PetId == modId {
				var nextAward int32
				if k+1 < len(petsInfo) {
					nextAward = petsInfo[k+1].Award
				}
				pet.NextAward = nextAward
				newPets = pet
				break
			}
		}
		petOther := this.GetIntroductionByModId(modId)
		newPets.MaxLevel = petOther.MaxLevel
		newPets.Name = petOther.Name
		newPets.Story = petOther.Story
		newPets.AwardTitle = petOther.AwardTitle
		//
		item := BagMgrSington.GetBagItemById(p.SnId, newPets.Fragment)
		if item != nil {
			newPets.HaveAmount = item.ItemNum
		}
		return newPets
	}
	return nil
}

// 商品人物总加成  人物功能变动需要修改
func (this *PetMgr) GetShopAward(shopInfo *ShopInfo, p *Player) (award int32) {
	roleGirl := this.GetRoleInfo(p, 2000001)
	if roleGirl != nil && roleGirl.Level > 0 /*|| !role.IsUsing*/ {
		//女孩加成
		if shopInfo.Ad > 0 && shopInfo.Type == Shop_Type_Coin && roleGirl.AwardType == 1 {
			award += roleGirl.Award
		}
	}
	roleBoy := this.GetRoleInfo(p, 2000002)
	if roleBoy != nil && roleBoy.Level > 0 /*|| !role.IsUsing*/ {
		//男孩加成
		if shopInfo.Ad == 0 && shopInfo.Type == Shop_Type_Coin && roleBoy.AwardType == 2 {
			award += roleBoy.Award
		}
	}
	return award
}

// 宠物加成  宠物功能变动需要修改
func (this *PetMgr) GetAwardPetByWelf(p *Player) (award int64) {
	petChick := this.GetPetInfo(p, 1000001)
	if petChick != nil && petChick.Level > 0 {
		award = int64(petChick.Award)
	}
	return
}
func (this *PetMgr) GetIntroductionByModId(modId int32) *RoleAndPetInfo {
	mod := srvdata.PBDB_Game_IntroductionMgr.GetData(modId)
	if mod == nil {
		return &RoleAndPetInfo{}
	}
	return &RoleAndPetInfo{
		ModId:      mod.Id,
		Name:       mod.Name,
		Story:      mod.Story,
		AwardTitle: mod.AwardTitle,
		MaxLevel:   mod.LevelMax,
	}
}
func (this *PetMgr) CheckShowRed(p *Player) {
	if p == nil {
		return
	}
	var roleRed, petRed bool
	//人物
	rolesInfo := srvdata.PBDB_Game_RoleMgr.Datas.Arr
	if rolesInfo != nil {
		needAmount := make(map[int32]int32)
		for k, roleInfo := range rolesInfo {
			if p.Roles != nil {
				level := p.Roles.ModUnlock[roleInfo.RoleId]
				if roleInfo.Level == level {
					if k+1 < len(rolesInfo) && roleInfo.RoleId == rolesInfo[k+1].RoleId {
						needAmount[roleInfo.Fragment] = rolesInfo[k+1].Amount
					}
				}
			}
		}
		for fragment, amount := range needAmount {
			item := BagMgrSington.GetBagItemById(p.SnId, fragment)
			if item != nil && item.ItemNum >= amount {
				roleRed = true
			}
		}
	}
	//宠物
	petsInfo := srvdata.PBDB_Game_PetMgr.Datas.Arr
	if petsInfo != nil {
		needAmount := make(map[int32]int32)
		for k, petInfo := range petsInfo {
			if p.Pets != nil {
				level := p.Pets.ModUnlock[petInfo.PetId]
				if petInfo.Level == level {
					if k+1 < len(petsInfo) && petInfo.PetId == petsInfo[k+1].PetId {
						needAmount[petInfo.Fragment] = petsInfo[k+1].Amount
					}
				}
			}
		}
		for fragment, amount := range needAmount {
			item := BagMgrSington.GetBagItemById(p.SnId, fragment)
			if item != nil && item.ItemNum >= amount {
				petRed = true
			}
		}
	}
	if roleRed {
		p.SendShowRed(hall_proto.ShowRedCode_Role, 0, 1)
	}
	if petRed {
		p.SendShowRed(hall_proto.ShowRedCode_Pet, 0, 1)
	}
}

//////////////////////////////////

func (this *PetMgr) Init() {

}
func (this *PetMgr) Update() {

}

func (this *PetMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(PetMgrSington, time.Second, 0)
}
