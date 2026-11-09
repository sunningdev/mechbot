package i18n

var vi = Text{
	DangerHead: `⚠️ MechBot: cảnh báo thanh lý

Giao thức: %s
Địa chỉ: %s
Health factor: %.4f (ngưỡng %.2f)
Tài sản thế chấp: %s
Khoản nợ: %s
`,
	ETHDrop:    "Sẽ bị thanh lý nếu ETH giảm thêm %.1f%% (giá thanh lý ETH %s)\n",
	DangerFoot: "\nVị thế có thể bị thanh lý khi health factor xuống dưới 1.0.",

	EarlyHead: `⏱ MechBot: cảnh báo sớm (Pro)

Giao thức: %s
Địa chỉ: %s
Health factor on-chain: %.4f
Ước tính theo giá Binance thời gian thực: %.4f (ngưỡng %.2f)
`,
	EarlyETH:  "ETH: %s → %s (%+.1f%%)\n",
	EarlyNote: "\nOracle on-chain thường chậm hơn sàn giao dịch; khi oracle cập nhật, vị thế của bạn sẽ xuống dưới ngưỡng. Bạn vẫn còn thời gian xử lý.\n",

	RateHigh: `📈 MechBot: cảnh báo lãi suất vay (Pro)

Giao thức: %s
Địa chỉ: %s
Lãi suất vay %s: %.2f%%/năm
Khoản nợ: %s, với mức này tốn thêm khoảng %s mỗi năm

Lãi suất tăng vọt thường có nghĩa là pool gần như đã bị vay hết. Vòng lặp (loop) của bạn có thể đang lỗ; hãy cân nhắc trả bớt nợ hoặc chuyển sang thị trường khác.`,
	BorrowWord: "khoản nợ",
	RateCalm: `✅ MechBot: lãi suất vay đã giảm

Giao thức: %s
Địa chỉ: %s
Lãi suất vay: %.2f%%/năm`,

	DepegSideCollateral:   "Tài sản thế chấp",
	DepegSideDebt:         "Khoản nợ",
	DepegEffectCollateral: "Tài sản thế chấp của bạn đang mất giá",
	DepegEffectDebt:       "Khoản nợ của bạn đang tăng",
	DepegEffectFavorable:  "Hiện tại điều này có lợi cho bạn, nhưng depeg thường đi kèm biến động mạnh",
	DepegEffectHF:         ", và health factor sẽ giảm khi oracle cập nhật",
	Depeg: `⚠️ MechBot: cảnh báo depeg (Pro)

Giao thức: %s
Địa chỉ: %s
%s %s (neo theo %s) hiện ở %.4f %s (lệch %+.2f%%)
Health factor: %.4f

%s. Hãy cân nhắc giảm đòn bẩy: trả bớt khoản vay hoặc thêm tài sản thế chấp khác.%s`,
	DepegCalm: `✅ MechBot: %s đã quay về gần mức neo

Giao thức: %s
Địa chỉ: %s
Hiện tại %.4f %s`,

	AppLine:     "Xử lý: %s",
	RescueHead:  "\nCách xử lý (đưa về %.2f):\n",
	RescueRepay: "· Trả %s\n",
	RescueAdd:   "· Hoặc thêm %s tài sản thế chấp\n",
	TokenApprox: "%s %s (khoảng %s)",

	SafeNoDebt: `✅ MechBot: vị thế đã an toàn

Giao thức: %s
Địa chỉ: %s

Khoản vay này đã được trả hết, không còn nợ.`,
	Safe: `✅ MechBot: vị thế đã an toàn

Giao thức: %s
Địa chỉ: %s
Health factor: %.4f (ngưỡng %.2f)
Tài sản thế chấp: %s
Khoản nợ: %s

Đã trở lại trên ngưỡng của bạn.`,

	ListHead:   "Địa chỉ %s\nNgưỡng %.2f\n",
	ListFailed: "\n%s\nTruy vấn thất bại: %v\n",
	ListPos:    "\n%s\nHealth factor %.4f\nThế chấp %s\nNợ %s\n",
	ListDrop:   "Thanh lý nếu ETH giảm %.1f%% (%s)\n",
	ListAPY:    "Lãi suất vay %.2f%%\n",
	ListQuiet:  "\nKhông có nợ: %s\n",
	ListSep:    ", ",

	AddedDebt:   "Đã thêm. Các vị thế hiện tại:\n\n",
	AddedFailed: "Đã thêm %s.\n\nMột số giao thức tạm thời chưa phản hồi. Hãy thử lại /list sau ít phút.",
	AddedQuiet: `Đã thêm %s.

Địa chỉ này hiện không có khoản vay nào: chúng tôi đã kiểm tra Aave V3, Compound V3 và Morpho Blue (Ethereum · Arbitrum · Optimism), nên không thể bị thanh lý.

Chúng tôi sẽ tiếp tục theo dõi. Nếu sau này địa chỉ vay trên các giao thức này, cảnh báo sẽ tự bật, không cần thêm lại.`,
	AddedStarter: "\n\nGói Starter chỉ cảnh báo cho Aave V3. Khoản vay trên Compound và Morpho cần gói Pro.",

	StarterLimits: `· Chỉ còn cảnh báo Aave V3; cảnh báo Compound và Morpho sẽ dừng
· Cảnh báo sớm, cảnh báo lãi suất và depeg sẽ dừng
· Bản đồ thanh lý bị trễ 1 giờ`,
	ProEndingPaid:  "Gói Pro của bạn sẽ hết hạn trong vòng 24 giờ",
	ProEndingTrial: "Gói dùng thử Pro 7 ngày của bạn sẽ hết hạn trong vòng 24 giờ",
	ProEnding: `⏳ MechBot: %s

Sau đó bạn sẽ chuyển sang Starter (miễn phí):
· Chỉ %d địa chỉ đầu tiên tiếp tục được theo dõi; số còn lại tạm dừng (không bị xoá)
%s

Giữ gói Pro: %s`,
	ProEndedPaid:   "Gói Pro đã hết hạn, bạn đã chuyển sang Starter",
	ProEndedTrial:  "Gói dùng thử Pro đã kết thúc, bạn đã chuyển sang Starter",
	ProEndedWatch:  "· %d địa chỉ đầu tiên vẫn được theo dõi",
	ProEndedPaused: "; %d địa chỉ còn lại tạm dừng (dữ liệu vẫn giữ, sẽ tiếp tục khi gia hạn)",
	ProEnded: `MechBot: %s

%s
%s

Gia hạn để lấy lại mọi thứ ngay: %s`,

	StatusStarter: "Starter (miễn phí vĩnh viễn, chỉ cảnh báo Aave V3)",
	StatusPro:     "Pro (mọi giao thức)",
	StatusTrial:   "Dùng thử Pro (mọi giao thức, hết hạn sẽ về Starter)",
	StatusNone:    "Không có",
	Status: `Tài khoản MechBot

Gói: %s
Địa chỉ: %d/%d
Pro đến: %s
Người mời: %s`,

	PlanMonth:   "Pro tháng",
	PlanQuarter: "Pro quý",
	PlanYear:    "Pro năm",
	PayTooLow:   "⚠️ Đã nhận %s USDT (%s)\n\nSố tiền thấp hơn gói rẻ nhất (%d USDT) nên Pro chưa được kích hoạt. Nếu có thắc mắc, hãy nhắn cho bot.",
	PayBonus:    "\n🎁 Thưởng mời bạn +%d ngày",
	PayDone:     "✅ Đã kích hoạt MechBot Pro\n\nĐã nhận %s USDT (%s): %s, %d ngày%s\nPro đến: %s",
	PayReferrer: "🎁 Thưởng mời bạn +%d ngày\n\nNgười bạn đã mời vừa mua %s lần đầu.\nPro đến: %s",

	Help: `MechBot: cảnh báo thanh lý

/add 0x... Theo dõi một địa chỉ (Starter 3, Pro 20)
/list Health factor trên từng giao thức
/remove 0x... Ngừng theo dõi một địa chỉ
/status Trạng thái tài khoản
/lang Đổi ngôn ngữ tin nhắn

Mỗi địa chỉ có ngưỡng mặc định 1.2. Bạn có thể đổi trong thẻ địa chỉ trên bảng điều khiển web.

Starter (miễn phí vĩnh viễn) cảnh báo cho Aave V3 (Ethereum, Arbitrum, Optimism). Pro thêm Compound V3 và Morpho Blue.`,
	HelpConsole:       "\n\nBảng điều khiển web: %s",
	TrialNote:         "\n\n🎁 Người dùng mới được %d ngày Pro: cảnh báo mọi giao thức, cảnh báo sớm và bản đồ thanh lý thời gian thực. Hết hạn không bị tính phí, bạn chỉ quay về gói Starter miễn phí.",
	UnknownCommand:    "Không có lệnh này.",
	RegisterFailed:    "Đăng ký thất bại. Vui lòng thử lại sau.",
	Registered:        "Đăng ký thành công.",
	AlreadyRegistered: "Bạn đã đăng ký rồi.",
	QueryFailed:       "Truy vấn thất bại. Vui lòng thử lại sau.",
	AddUsage:          "Cách dùng: /add kèm địa chỉ ví bắt đầu bằng 0x",
	AddFailed:         "Thêm thất bại. Vui lòng thử lại sau.",
	AddExists:         "Địa chỉ này đã được theo dõi.",
	AddLimit:          "Gói của bạn cho phép tối đa %d địa chỉ.",
	RemoveUsage:       "Cách dùng: /remove kèm địa chỉ ví bắt đầu bằng 0x",
	Removed:           "Đã xoá %s",
	RemoveMissing:     "Địa chỉ này không có trong danh sách của bạn.",
	RemoveFailed:      "Xoá thất bại. Vui lòng thử lại sau.",
	ThresholdMoved:    "Ngưỡng giờ được đặt riêng cho từng địa chỉ (mặc định 1.2). Hãy đổi trong thẻ địa chỉ trên bảng điều khiển web.",
	NoAddresses:       "Chưa có địa chỉ nào. Thêm bằng /add 0x...",
	Paused:            "\n\n(Vượt giới hạn Starter, cảnh báo tạm dừng)",

	LoginPrivate: "Hãy xác nhận đăng nhập trong cuộc trò chuyện riêng với bot.",
	LoginInvalid: "Liên kết đăng nhập không hợp lệ hoặc đã hết hạn. Hãy quay lại trang web và bấm lại “Đăng nhập bằng Telegram”.",
	LoginPrompt: `Bạn đang đăng nhập bảng điều khiển web của MechBot.

Mã hiển thị trên trang web: %s

Chỉ bấm “Xác nhận đăng nhập” nếu hai mã trùng nhau. Nếu bạn không vừa đăng nhập trên trang web, đừng xác nhận.`,
	LoginButton:       "Xác nhận đăng nhập",
	LoginExpiredToast: "Liên kết đã hết hạn",
	LoginExpired:      "Liên kết đăng nhập đã hết hạn. Hãy quay lại trang web và bấm lại “Đăng nhập bằng Telegram”.",
	LoginFailedToast:  "Đăng nhập thất bại, hãy thử lại",
	LoginUsedToast:    "Liên kết không còn hiệu lực",
	LoginUsed:         "Liên kết đăng nhập này đã được dùng hoặc đã hết hạn.",
	LoginConfirmToast: "Đã xác nhận",
	LoginConfirmed:    "Đã xác nhận đăng nhập. Hãy quay lại trang web, bảng điều khiển sẽ tự mở.",

	LangPrompt: "Chọn ngôn ngữ tin nhắn của bot:",
	LangSet:    "Xong. Từ giờ tin nhắn sẽ được gửi bằng tiếng Việt.",
}
