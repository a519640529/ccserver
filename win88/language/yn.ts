const win = window as any;

export let languages = {
    // Data
    "DBGameItem": {
        "item_10001_describe": "Dùng tăng cấp, mở khoá Pet Gà Đá; dùng ở giao diện tính năng nhân vật. \nCách nhận: Điểm Danh 14 Ngày, rớt trong TienLen ",
        "item_10001_name": "Mảnh Gà Đá",
        "item_20001_describe": "Cách nhận: Điểm Danh 14 Ngày, rớt trong TienLen",
        "item_20001_name": "Mảnh Kupa",
        "item_20002_describe": "Tác dụng: Tăng cấp, mở khóa nhân vật Butterfly；dùng ở giao diện tính năng nhân vật. \nCách nhận: Điểm Danh 14 Ngày, rớt trong TienLen",
        "item_20002_name": "Mảnh Butterfly",
        "item_30001_describe": "Cách dùng: Đổi vật phẩm trong Shop Đổi.\n Cách nhận: Có tỷ lệ rớt trong Tienlen; tiền cược càng cao, tỷ lệ rớt và số lượng càng cao",
        "item_30001_name": "Thẻ V",
        "item_40001_describe": "Chức năng: được sử dụng để đăng ký các sự kiện vô địch;\nLộ trình đầu ra: Tham gia bất kỳ giải đấu nào và giành chức vô địch",
        "item_40001_name": "vé vô địch",
    },
    "DBGameIntroduction": {
        "Introduction_1000001_AwardTitle": "Tăng thưởng hỗ trợ phá sản",
        "Introduction_1000001_name": "Gà Đá",
        "Introduction_1000001_story": "Chim cu. . . chim cu",
        "Introduction_2000001_AwardTitle": "Thưởng tiền thưởng khi xem video trong trung tâm thương mại",
        "Introduction_2000001_name": "Butterfly",
        "Introduction_2000001_story": "Cô gái yêu thiên nhiên, yêu hòa bình",
        "Introduction_2000002_AwardTitle": "Mua Vàng ở Shop tăng thưởng",
        "Introduction_2000002_name": "Kupa",
        "Introduction_2000002_story": "Chàng trai vùng biển cởi mở, thích bơi lội, lướt sóng",
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
