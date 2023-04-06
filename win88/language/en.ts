const win = window as any;

export let languages = {
    // Data
    "DBGameItem": {
        "item_10001_describe": "Function: Used to upgrade and unlock pet Gu; it can be used on the character function page.\nOutput: Signed in on the 14th, dropped in TienLen's gameplay",
        "item_10001_name": "Cockfight Pieces",
        "item_20001_describe": "Function: Used to upgrade and unlock the character Kupa; it can be used on the character function page.\nOutput: Signed in on the 14th, dropped in TienLen's gameplay",
        "item_20001_name": "Kupa pieces",
        "item_20002_describe": "Function: Used to upgrade and unlock the character Tim; it can be used on the character function page.\nOutput: Signed in on the 14th, dropped in TienLen's gameplay",
        "item_20002_name": "Dgirls pieces",
        "item_30001_describe": "Function: It can be used in the exchange mall to exchange for the item you want.\nOutput: The probability of falling in Tienlen gameplay; the higher the ante, the higher the probability and number of drops",
        "item_30001_name": "V card",
        "item_40001_describe": "Function: used to sign up for championship events;\nOutput path: Participate in any tournament and win the championship",
        "item_40001_name": "Championship tickets",
    },
    "DBGameIntroduction": {
        "Introduction_1000001_AwardTitle": "Bonus for bankruptcy subsidy",
        "Introduction_1000001_name": "Cockfight",
        "Introduction_1000001_story": "Cuckoo. . . cuckoo",
        "Introduction_2000001_AwardTitle": "Bonus for watching videos in the mall",
        "Introduction_2000001_name": "Dgirls",
        "Introduction_2000001_story": "A dream girl who loves nature and peace",
        "Introduction_2000002_AwardTitle": "Bonus for purchasing gold coins",
        "Introduction_2000002_name": "Kupa",
        "Introduction_2000002_story": "Growing up on the beach, he is fond of water and surfing. He is a sunny beach boy",
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
