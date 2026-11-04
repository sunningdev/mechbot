import type { Post } from "./posts";

export const viPosts: Post[] = [
  {
    slug: "health-factor",
    title: "Hệ số sức khỏe được tính thế nào? Hiểu con số “còn tới thanh lý”",
    summary:
      "Aave, Compound và Morpho gọi ngưỡng thanh lý bằng những cái tên khác nhau, nhưng bản chất là một. Một công thức giải thích tất cả, và cho bạn biết giá còn có thể giảm bao nhiêu.",
    date: "2026-10-28",
    tag: "Nguyên lý",
    readMin: 5,
    body: [
      {
        type: "p",
        text: "Giao thức cho vay không quan tâm bạn vay bao nhiêu. Nó chỉ quan tâm một điều: sau khi chiết khấu tài sản thế chấp theo ngưỡng thanh lý, phần còn lại có đủ bù khoản nợ không? Hệ số sức khỏe (Health Factor, HF) chính là tỷ lệ đó.",
      },
      { type: "h", text: "Một công thức" },
      { type: "code", text: "HF = Σ(giá trị thế chấp × ngưỡng thanh lý) / giá trị nợ" },
      {
        type: "p",
        text: "Lớn hơn 1 là an toàn. Bằng 1 là chạm ngưỡng thanh lý. Dưới 1, bất kỳ ai cũng có thể thanh lý vị thế của bạn và lấy một phần tài sản thế chấp làm phần thưởng.",
      },
      { type: "h", text: "Ba giao thức, ba cách gọi" },
      {
        type: "list",
        items: [
          "Aave V3: mỗi tài sản có một Liquidation Threshold, và hợp đồng trả về hệ số sức khỏe trực tiếp.",
          "Compound V3: mỗi tài sản thế chấp có một liquidateCollateralFactor; chúng tôi tự áp dụng cùng công thức.",
          "Morpho Blue: mỗi thị trường có một LLTV (tỷ lệ vay trên giá trị khi thanh lý). Cùng công thức, chỉ với một loại tài sản thế chấp.",
        ],
      },
      { type: "h", text: "“Còn tới thanh lý” nghĩa là gì" },
      {
        type: "p",
        text: "Trong bảng điều khiển, “còn tới thanh lý” là mức giá tài sản thế chấp còn có thể giảm, khi khoản nợ không đổi, trước khi bị thanh lý:",
      },
      { type: "code", text: "Còn tới thanh lý = 1 − 1 / HF" },
      {
        type: "p",
        text: "Với HF 1.25, tài sản thế chấp còn có thể giảm 20%. Với HF 1.07, chỉ còn khoảng 6.5% — một cây nến đỏ lớn là đủ.",
      },
      {
        type: "note",
        text: "Nếu bạn vay tài sản biến động (ví dụ vay ETH), khoản nợ cũng thay đổi, nên khoảng đệm thực tế sẽ khác. Hãy xem con số này là tham khảo trực quan, không phải dự đoán chính xác.",
      },
      { type: "h", text: "Nên đặt ngưỡng cảnh báo bao nhiêu" },
      {
        type: "p",
        text: "Thế chấp stablecoin để vay stablecoin có thể đặt thấp hơn, khoảng 1.1. Thế chấp ETH hoặc BTC để vay stablecoin, nên đặt 1.3 đến 1.5 để có thời gian nạp thêm.",
      },
    ],
  },
  {
    slug: "rpc-failover",
    title: "Node chết thì sao? Thiết kế chuyển đổi đa node của MechBot",
    summary:
      "RPC công khai có thể giới hạn, quá thời gian và ngừng hoạt động. Đây là cách cảnh báo vẫn được gửi khi node gặp sự cố — và vì sao “không đọc được” không bao giờ được coi là “an toàn”.",
    date: "2026-10-26",
    tag: "Kỹ thuật",
    readMin: 6,
    body: [
      {
        type: "p",
        text: "Lỗi tệ nhất của cảnh báo thanh lý không phải là báo lỗi, mà là im lặng: node chết, không đọc được gì, không gửi gì, và người dùng tưởng mọi thứ vẫn ổn. Thiết kế chuyển đổi của MechBot xoay quanh một nguyên tắc: dữ liệu không đọc được thì không bao giờ được coi là an toàn.",
      },
      { type: "h", text: "Mỗi chuỗi có một nhóm node" },
      {
        type: "p",
        text: "Mỗi chuỗi có nhiều RPC: các node bạn cấu hình đứng trước, sau đó là các node dự phòng công khai có sẵn. Mỗi truy vấn thử lần lượt, ghi nhớ node nào thành công và ưu tiên dùng nó lần sau.",
      },
      { type: "code", text: "Node cấu hình → dự phòng 1 → dự phòng 2 → …\nThành công: ghi nhớ node này\nThất bại: tự động chuyển sang node tiếp theo" },
      { type: "h", text: "Khi tất cả node đều thất bại" },
      {
        type: "list",
        items: [
          "Vị thế được đánh dấu “truy vấn thất bại, thử lại vòng sau” trong bảng điều khiển — không bao giờ hiển thị là an toàn.",
          "Nếu một nguồn dữ liệu thất bại với mọi địa chỉ trong nhiều vòng liên tiếp, cảnh báo trực sẽ gửi tới Telegram của người vận hành.",
          "Khi node hồi phục, một thông báo hồi phục xác nhận đường truyền đã ổn định trở lại.",
        ],
      },
      { type: "h", text: "Bản thân chương trình cũng phải trụ vững" },
      {
        type: "p",
        text: "Việc đọc giao thức, vòng kiểm tra, quét thanh toán và xử lý tin nhắn bot đều có cơ chế chống sập riêng. Dữ liệu lỗi từ một vị thế không thể làm hỏng cả vòng kiểm tra, càng không thể làm dừng dịch vụ.",
      },
      {
        type: "note",
        text: "Khi API Telegram trả về lỗi, token của bot sẽ được ẩn khỏi log để bí mật không bao giờ lọt vào hệ thống ghi log.",
      },
    ],
  },
  {
    slug: "no-spam",
    title: "Vì sao MechBot không spam: vòng đời của một cảnh báo",
    summary:
      "Hàng trăm thông báo mỗi ngày cũng như không có. Chúng tôi chỉ gọi bạn khi vượt ngưỡng cảnh báo, và không lặp lại trong vòng một giờ cho cùng một vị thế.",
    date: "2026-10-24",
    tag: "Sản phẩm",
    readMin: 4,
    body: [
      {
        type: "p",
        text: "Nhiều bot theo dõi đẩy mọi biến động giá. Người dùng tắt thông báo, rồi bỏ lỡ đúng cảnh báo quan trọng nhất. MechBot hướng tới điều ngược lại: bình thường im lặng, chỉ đánh thức bạn trước khi bị thanh lý.",
      },
      { type: "h", text: "Kiểm tra: mỗi 5 phút" },
      {
        type: "p",
        text: "Mỗi 5 phút, hệ thống đọc vị thế của tất cả địa chỉ của bạn trên từng giao thức, trực tiếp từ hợp đồng on-chain, không phụ thuộc API dữ liệu bên thứ ba.",
      },
      { type: "h", text: "Quyết định: chỉ nhìn một ngưỡng" },
      {
        type: "list",
        items: [
          "Không có khoản vay, hoặc hệ số sức khỏe trên ngưỡng cảnh báo: không gửi gì.",
          "Xuống dưới ngưỡng: nhắn riêng ngay, kèm giao thức, chuỗi, hệ số sức khỏe, thế chấp và nợ.",
          "Vẫn dưới ngưỡng: không nhắc lại trong 1 giờ; sau 1 giờ nếu chưa hồi phục, nhắc thêm một lần.",
          "Trở lại trên ngưỡng: một tin “đã an toàn trở lại”, rồi im lặng như cũ.",
        ],
      },
      { type: "code", text: "An toàn ──giảm──▶ Cảnh báo (nhắn riêng)\n  ▲                  │ im lặng 1 giờ\n  └──hồi phục── Đã an toàn trở lại (một tin)" },
      { type: "h", text: "Kết quả" },
      {
        type: "p",
        text: "Một vị thế đi từ an toàn tới nguy hiểm rồi hồi phục thường chỉ tốn của bạn hai tin nhắn. Mỗi tin đều đáng để bạn dừng việc đang làm và xem ngay.",
      },
    ],
  },
];
