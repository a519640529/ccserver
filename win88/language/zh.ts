const win = window as any;

export let languages = {
    // Data
    "DBGameItem": {
        "item_10001_describe": "作用：用于升级，解锁宠物斗鸡；可在角色功能页面使用。\\n产出途径：14日签到，TienLen玩法中掉落\n",
        "item_10001_name": "斗鸡碎片",
        "item_20001_describe": "作用：用于升级，解锁角色酷帕；可在角色功能页面使用。\\n产出途径：14日签到，TienLen玩法中掉落",
        "item_20001_name": "酷帕碎片",
        "item_20002_describe": "作用：用于升级，解锁角色蝶女；可在角色功能页面使用。\\n产出途径：14日签到，TienLen玩法中掉落",
        "item_20002_name": "蝶女碎片",
        "item_30001_describe": "作用：可在兑换商城中使用，兑换您想要的物品。\\n产出途径：Tienlen玩法中概率掉落；底注越高，掉落概率和数量越高",
        "item_30001_name": "V卡",
        "item_40001_describe": "用途：用于报名冠军赛事；\n产出途径：参加任意锦标赛，获得冠军",
        "item_40001_name": "冠军赛门票",
    },
    "DBGameIntroduction": {
        "Introduction_1000001_AwardTitle": "破产补助奖励加成",
        "Introduction_1000001_name": "斗鸡",
        "Introduction_1000001_story": "咕咕。。。咕咕",
        "Introduction_2000001_AwardTitle": "商城观看视频奖励加成",
        "Introduction_2000001_name": "蝶女",
        "Introduction_2000001_story": "一个热爱大自然，热爱和平的追梦女孩",
        "Introduction_2000002_AwardTitle": "商城购买金币奖励加成",
        "Introduction_2000002_name": "酷帕",
        "Introduction_2000002_story": "从小在海边长大，习水性，爱好冲浪，一个阳光开朗的沙滩男孩",
    },
};

if (!win.languages) {
    win.languages = {};
}
if (win.languages.zh) {
    for (let item in languages) {
        win.languages.zh[item] = languages[item];
    }
} else {
    win.languages.zh = languages;
}
