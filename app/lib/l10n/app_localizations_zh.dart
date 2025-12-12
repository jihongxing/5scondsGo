// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => '5秒开奖';

  @override
  String get login => '登录';

  @override
  String get register => '注册';

  @override
  String get username => '用户名';

  @override
  String get password => '密码';

  @override
  String get inviteCode => '邀请码';

  @override
  String get loginButton => '登录';

  @override
  String get registerButton => '注册';

  @override
  String get noAccount => '没有账号？';

  @override
  String get hasAccount => '已有账号？';

  @override
  String get home => '首页';

  @override
  String get rooms => '房间';

  @override
  String get profile => '我的';

  @override
  String get wallet => '钱包';

  @override
  String get joinRoom => '加入房间';

  @override
  String get leaveRoom => '离开房间';

  @override
  String get createRoom => '创建房间';

  @override
  String get roomName => '房间名称';

  @override
  String get betAmount => '下注金额';

  @override
  String get winnerCount => '赢家数量';

  @override
  String get maxPlayers => '最大人数';

  @override
  String get commissionRate => '抽成比例';

  @override
  String get phaseWaiting => '等待中';

  @override
  String get phaseCountdown => '倒计时';

  @override
  String get phaseBetting => '下注中';

  @override
  String get phaseInGame => '游戏中';

  @override
  String get phaseSettlement => '结算中';

  @override
  String get phaseReset => '重置中';

  @override
  String get autoReady => '自动准备';

  @override
  String get ready => '已准备';

  @override
  String get notReady => '未准备';

  @override
  String get balance => '余额';

  @override
  String get frozenBalance => '冻结';

  @override
  String get deposit => '充值';

  @override
  String get withdraw => '提现';

  @override
  String get currentRound => '第几轮';

  @override
  String get poolAmount => '奖池';

  @override
  String get winners => '赢家';

  @override
  String get prize => '奖金';

  @override
  String get players => '玩家';

  @override
  String get online => '在线';

  @override
  String get offline => '离线';

  @override
  String get submit => '提交';

  @override
  String get cancel => '取消';

  @override
  String get confirm => '确认';

  @override
  String get close => '关闭';

  @override
  String get save => '保存';

  @override
  String get error => '错误';

  @override
  String get success => '成功';

  @override
  String get loading => '加载中...';

  @override
  String get noData => '暂无数据';

  @override
  String get invalidCredentials => '用户名或密码错误';

  @override
  String get usernameExists => '用户名已存在';

  @override
  String get invalidInviteCode => '邀请码无效';

  @override
  String get roomFull => '房间已满';

  @override
  String get insufficientBalance => '余额不足';

  @override
  String get congratulations => '恭喜！';

  @override
  String get youWon => '你赢了！';

  @override
  String get betterLuckNextTime => '下次好运';

  @override
  String get verifyResult => '验证结果';

  @override
  String get commitHash => '承诺哈希';

  @override
  String get revealSeed => '随机种子';

  @override
  String get roomThemeTitle => '房间主题';

  @override
  String get roomThemeOwnerOnly => '只有房主可以更改主题';

  @override
  String get themeClassic => '经典';

  @override
  String get themeNeon => '霓虹';

  @override
  String get themeOcean => '海洋';

  @override
  String get themeForest => '森林';

  @override
  String get themeLuxury => '奢华';

  @override
  String get settings => '设置';

  @override
  String get language => '语言';

  @override
  String get logout => '退出登录';

  @override
  String get leaderboard => '排行榜';

  @override
  String get friends => '好友';

  @override
  String get spectate => '观战';

  @override
  String get switchToParticipant => '加入游戏';

  @override
  String get betPerRound => '单轮下注';

  @override
  String get gameHistory => '游戏记录';

  @override
  String minPlayersToStart(int count) {
    return '最少$count人开始';
  }

  @override
  String get congratsWin => '恭喜获胜！';

  @override
  String get wonPrize => '赢得';

  @override
  String playerWon(String names) {
    return '$names 获胜';
  }

  @override
  String bettingComplete(int count, String pool) {
    return '下注完成：$count人参与，奖池 ¥$pool';
  }

  @override
  String get noWinner => '无人中奖';

  @override
  String wonAmount(String names, String prize) {
    return '$names 赢得 ¥$prize';
  }

  @override
  String get gameCancelled => '游戏取消';

  @override
  String get notEnoughPlayers => '人数不足';

  @override
  String get insufficientBalanceDisqualified => '余额不足，无法参与本轮游戏';

  @override
  String get disqualifiedFromRound => '您已被取消本轮参与资格';

  @override
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers) {
    return '回合取消：需要$minRequired人，仅$currentPlayers人符合条件';
  }

  @override
  String get gameRecords => '游戏记录';

  @override
  String get myProfile => '我的';

  @override
  String get owner => '房主';

  @override
  String get admin => '管理员';

  @override
  String get player => '玩家';

  @override
  String get commission => '佣金';

  @override
  String get noRooms => '暂无房间';

  @override
  String get contactOwnerToCreate => '联系房主创建房间';

  @override
  String get enterRoomPassword => '请输入房间密码';

  @override
  String get passwordHint => '密码';

  @override
  String get availableBalance => '可用余额';

  @override
  String get totalBalance => '总余额';

  @override
  String get marginDeposit => '保证金';

  @override
  String get commissionEarnings => '佣金收益';

  @override
  String get myInviteCode => '我的邀请码';

  @override
  String get inviteCodeCopied => '邀请码已复制';

  @override
  String get selectLanguage => '选择语言';

  @override
  String get confirmLogout => '确认登出';

  @override
  String get confirmLogoutMessage => '确定要退出登录吗？';

  @override
  String get loadFailed => '加载失败';

  @override
  String get retry => '重试';

  @override
  String get noFriends => '暂无好友';

  @override
  String get clickToAddFriend => '点击右上角添加好友';

  @override
  String get searchFriends => '搜索好友...';

  @override
  String get addFriend => '添加好友';

  @override
  String get userId => '用户ID';

  @override
  String get enterUserId => '请输入用户ID';

  @override
  String get send => '发送';

  @override
  String get friendRequestSent => '好友请求已发送';

  @override
  String get sendFailed => '发送失败';

  @override
  String get inRoom => '在房间中';

  @override
  String get inviteToRoom => '邀请加入房间';

  @override
  String get removeFriend => '删除好友';

  @override
  String get confirmRemoveFriend => '删除好友';

  @override
  String confirmRemoveFriendMessage(String name) {
    return '确定要删除好友 $name 吗？';
  }

  @override
  String get remove => '删除';

  @override
  String get totalGames => '总场次';

  @override
  String get wins => '胜场';

  @override
  String get winRate => '胜率';

  @override
  String get totalBet => '总投注';

  @override
  String get totalWon => '总赢得';

  @override
  String get netProfit => '净盈亏';

  @override
  String roundNumber(int number) {
    return '第$number轮';
  }

  @override
  String get skipped => '跳过';

  @override
  String get bet => '投注';

  @override
  String get ownerWallet => '房主钱包';

  @override
  String get marginBalance => '保证金';

  @override
  String get commissionBalance => '佣金收益';

  @override
  String get totalCommission => '累计佣金';

  @override
  String get canRechargePlayer => '可给玩家充值';

  @override
  String get fixedGuarantee => '固定担保';

  @override
  String get canTransferToBalance => '可转入余额';

  @override
  String get transferToBalance => '收益转余额';

  @override
  String get transferAmount => '转账金额';

  @override
  String maxTransfer(String amount) {
    return '最多 $amount';
  }

  @override
  String get transfer => '转入';

  @override
  String get transferNote => '说明：将佣金收益转入可用余额后，可用于给玩家充值或申请提现';

  @override
  String get submitFundRequest => '发起资金申请';

  @override
  String get ownerDeposit => '余额充值';

  @override
  String get ownerWithdraw => '余额提现';

  @override
  String get marginDeposit2 => '保证金充值';

  @override
  String get amount => '金额';

  @override
  String get remarkOptional => '备注（可选）';

  @override
  String get submitRequest => '提交申请';

  @override
  String get fundRequestNote => '说明：充值/提现申请提交后，需要管理员在后台审核通过后，余额才会实际变动。';

  @override
  String get fundRequestRecords => '资金申请记录';

  @override
  String get pending => '待审核';

  @override
  String get approved => '已通过';

  @override
  String get rejected => '已拒绝';

  @override
  String get transactionHistory => '资金流水';

  @override
  String get gameBet => '游戏下注';

  @override
  String get gameWin => '游戏获胜';

  @override
  String get gameRefund => '游戏退款';

  @override
  String get ownerCommission => '房主佣金';

  @override
  String get platformShare => '平台抽成';

  @override
  String get pleaseEnterAmount => '请输入金额';

  @override
  String get invalidAmountFormat => '金额格式不正确';

  @override
  String get pleaseSelectType => '请选择申请类型';

  @override
  String get fundRequestSubmitted => '资金申请已提交，等待审核';

  @override
  String get submitFailed => '提交失败';

  @override
  String get pleaseEnterTransferAmount => '请输入转账金额';

  @override
  String transferAmountExceeds(String amount) {
    return '转账金额不能超过 ¥$amount';
  }

  @override
  String get transferSuccess => '转账成功';

  @override
  String get transferFailed => '转账失败';

  @override
  String get pendingPlayerRequests => '待审批的玩家申请';

  @override
  String get noPendingRequests => '暂无待审批的申请';

  @override
  String get unknownUser => '未知用户';

  @override
  String get confirmApprove => '确认通过';

  @override
  String get confirmReject => '确认拒绝';

  @override
  String get confirmApproveMessage => '确定要通过这个申请吗？';

  @override
  String get confirmRejectMessage => '确定要拒绝这个申请吗？';

  @override
  String get approve => '通过';

  @override
  String get reject => '拒绝';

  @override
  String get requestApproved => '申请已通过';

  @override
  String get requestRejected => '申请已拒绝';

  @override
  String get operationFailed => '操作失败';

  @override
  String get basicInfo => '基本信息';

  @override
  String get gameSettings => '游戏设置';

  @override
  String get feeSettings => '费率设置';

  @override
  String get securitySettings => '安全设置';

  @override
  String get platformFeeFixed => '平台费率（固定）';

  @override
  String get ownerFee => '房主费率';

  @override
  String get totalFee => '总费率';

  @override
  String get adjustOwnerFee => '调整房主费率:';

  @override
  String get feeNote => '说明：每局游戏结算时，从奖池中扣除总费率作为平台和房主收益';

  @override
  String get roomPassword => '房间密码';

  @override
  String get setPassword => '设置密码';

  @override
  String get pleaseEnterRoomName => '请输入房间名称';

  @override
  String get pleaseEnterPassword => '请输入密码';

  @override
  String get createFailed => '创建失败';

  @override
  String get winnersMustLessThanPlayers => '获胜人数必须小于最大玩家数';

  @override
  String minPlayersHint(int count) {
    return '每局游戏的获胜人数，最少需要 $count 人才能开始游戏';
  }

  @override
  String get maxPlayersHint => '房间最多容纳的玩家数';

  @override
  String get connectionError => '连接错误';

  @override
  String get connectionClosed => '连接已断开，正在重连...';

  @override
  String get joinRoomFailed => '连接房间失败';

  @override
  String get joinRoomTitle => '加入房间';

  @override
  String get currentBalance => '当前余额';

  @override
  String get minBalanceRequired => '最低余额要求';

  @override
  String get insufficientBalanceCannotJoin => '余额不足，无法加入房间';

  @override
  String get riskWarning => '风险提示';

  @override
  String get riskWarningTip1 => '游戏有风险，请理性参与';

  @override
  String get riskWarningTip2 => '进入房间后将自动参与每轮游戏';

  @override
  String get confirmJoin => '确认加入';

  @override
  String get minPlayersRequired => '最少开始人数';

  @override
  String get winnersPerRound => '每轮赢家数';

  @override
  String get maxPlayersAllowed => '最大参与人数';

  @override
  String get betPerRoundLabel => '每轮下注';

  @override
  String get personUnit => '人';
}

/// The translations for Chinese, as used in Taiwan (`zh_TW`).
class AppLocalizationsZhTw extends AppLocalizationsZh {
  AppLocalizationsZhTw() : super('zh_TW');

  @override
  String get appTitle => '5秒開獎';

  @override
  String get login => '登入';

  @override
  String get register => '註冊';

  @override
  String get username => '用戶名';

  @override
  String get password => '密碼';

  @override
  String get inviteCode => '邀請碼';

  @override
  String get loginButton => '登入';

  @override
  String get registerButton => '註冊';

  @override
  String get noAccount => '沒有帳號？';

  @override
  String get hasAccount => '已有帳號？';

  @override
  String get home => '首頁';

  @override
  String get rooms => '房間';

  @override
  String get profile => '我的';

  @override
  String get wallet => '錢包';

  @override
  String get joinRoom => '加入房間';

  @override
  String get leaveRoom => '離開房間';

  @override
  String get createRoom => '創建房間';

  @override
  String get roomName => '房間名稱';

  @override
  String get betAmount => '下注金額';

  @override
  String get winnerCount => '贏家數量';

  @override
  String get maxPlayers => '最大人數';

  @override
  String get commissionRate => '抽成比例';

  @override
  String get phaseWaiting => '等待中';

  @override
  String get phaseCountdown => '倒數計時';

  @override
  String get phaseBetting => '下注中';

  @override
  String get phaseInGame => '遊戲中';

  @override
  String get phaseSettlement => '結算中';

  @override
  String get phaseReset => '重置中';

  @override
  String get autoReady => '自動準備';

  @override
  String get ready => '已準備';

  @override
  String get notReady => '未準備';

  @override
  String get balance => '餘額';

  @override
  String get frozenBalance => '凍結';

  @override
  String get deposit => '充值';

  @override
  String get withdraw => '提現';

  @override
  String get currentRound => '第幾輪';

  @override
  String get poolAmount => '獎池';

  @override
  String get winners => '贏家';

  @override
  String get prize => '獎金';

  @override
  String get players => '玩家';

  @override
  String get online => '在線';

  @override
  String get offline => '離線';

  @override
  String get submit => '提交';

  @override
  String get cancel => '取消';

  @override
  String get confirm => '確認';

  @override
  String get close => '關閉';

  @override
  String get save => '保存';

  @override
  String get error => '錯誤';

  @override
  String get success => '成功';

  @override
  String get loading => '載入中...';

  @override
  String get noData => '暫無數據';

  @override
  String get invalidCredentials => '用戶名或密碼錯誤';

  @override
  String get usernameExists => '用戶名已存在';

  @override
  String get invalidInviteCode => '邀請碼無效';

  @override
  String get roomFull => '房間已滿';

  @override
  String get insufficientBalance => '餘額不足';

  @override
  String get congratulations => '恭喜！';

  @override
  String get youWon => '你贏了！';

  @override
  String get betterLuckNextTime => '下次好運';

  @override
  String get verifyResult => '驗證結果';

  @override
  String get commitHash => '承諾哈希';

  @override
  String get revealSeed => '隨機種子';

  @override
  String get roomThemeTitle => '房間主題';

  @override
  String get roomThemeOwnerOnly => '只有房主可以更改主題';

  @override
  String get themeClassic => '經典';

  @override
  String get themeNeon => '霓虹';

  @override
  String get themeOcean => '海洋';

  @override
  String get themeForest => '森林';

  @override
  String get themeLuxury => '奢華';

  @override
  String get settings => '設定';

  @override
  String get language => '語言';

  @override
  String get logout => '登出';

  @override
  String get leaderboard => '排行榜';

  @override
  String get friends => '好友';

  @override
  String get spectate => '觀戰';

  @override
  String get switchToParticipant => '加入遊戲';

  @override
  String get betPerRound => '單輪下注';

  @override
  String get gameHistory => '遊戲記錄';

  @override
  String minPlayersToStart(int count) {
    return '最少$count人開始';
  }

  @override
  String get congratsWin => '恭喜獲勝！';

  @override
  String get wonPrize => '贏得';

  @override
  String playerWon(String names) {
    return '$names 獲勝';
  }

  @override
  String bettingComplete(int count, String pool) {
    return '下注完成：$count人參與，獎池 ¥$pool';
  }

  @override
  String get noWinner => '無人中獎';

  @override
  String wonAmount(String names, String prize) {
    return '$names 贏得 ¥$prize';
  }

  @override
  String get gameCancelled => '遊戲取消';

  @override
  String get notEnoughPlayers => '人數不足';

  @override
  String get insufficientBalanceDisqualified => '餘額不足，無法參與本輪遊戲';

  @override
  String get disqualifiedFromRound => '您已被取消本輪參與資格';

  @override
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers) {
    return '回合取消：需要$minRequired人，僅$currentPlayers人符合條件';
  }

  @override
  String get gameRecords => '遊戲記錄';

  @override
  String get myProfile => '我的';

  @override
  String get owner => '房主';

  @override
  String get admin => '管理員';

  @override
  String get player => '玩家';

  @override
  String get commission => '佣金';

  @override
  String get noRooms => '暫無房間';

  @override
  String get contactOwnerToCreate => '聯繫房主創建房間';

  @override
  String get enterRoomPassword => '請輸入房間密碼';

  @override
  String get passwordHint => '密碼';

  @override
  String get availableBalance => '可用餘額';

  @override
  String get totalBalance => '總餘額';

  @override
  String get marginDeposit => '保證金';

  @override
  String get commissionEarnings => '佣金收益';

  @override
  String get myInviteCode => '我的邀請碼';

  @override
  String get inviteCodeCopied => '邀請碼已複製';

  @override
  String get selectLanguage => '選擇語言';

  @override
  String get confirmLogout => '確認登出';

  @override
  String get confirmLogoutMessage => '確定要退出登錄嗎？';

  @override
  String get loadFailed => '載入失敗';

  @override
  String get retry => '重試';

  @override
  String get noFriends => '暫無好友';

  @override
  String get clickToAddFriend => '點擊右上角添加好友';

  @override
  String get searchFriends => '搜尋好友...';

  @override
  String get addFriend => '添加好友';

  @override
  String get userId => '用戶ID';

  @override
  String get enterUserId => '請輸入用戶ID';

  @override
  String get send => '發送';

  @override
  String get friendRequestSent => '好友請求已發送';

  @override
  String get sendFailed => '發送失敗';

  @override
  String get inRoom => '在房間中';

  @override
  String get inviteToRoom => '邀請加入房間';

  @override
  String get removeFriend => '刪除好友';

  @override
  String get confirmRemoveFriend => '刪除好友';

  @override
  String confirmRemoveFriendMessage(String name) {
    return '確定要刪除好友 $name 嗎？';
  }

  @override
  String get remove => '刪除';

  @override
  String get totalGames => '總場次';

  @override
  String get wins => '勝場';

  @override
  String get winRate => '勝率';

  @override
  String get totalBet => '總投注';

  @override
  String get totalWon => '總贏得';

  @override
  String get netProfit => '淨盈虧';

  @override
  String roundNumber(int number) {
    return '第$number輪';
  }

  @override
  String get skipped => '跳過';

  @override
  String get bet => '投注';

  @override
  String get ownerWallet => '房主錢包';

  @override
  String get marginBalance => '保證金';

  @override
  String get commissionBalance => '佣金收益';

  @override
  String get totalCommission => '累計佣金';

  @override
  String get canRechargePlayer => '可給玩家充值';

  @override
  String get fixedGuarantee => '固定擔保';

  @override
  String get canTransferToBalance => '可轉入餘額';

  @override
  String get transferToBalance => '收益轉餘額';

  @override
  String get transferAmount => '轉帳金額';

  @override
  String maxTransfer(String amount) {
    return '最多 $amount';
  }

  @override
  String get transfer => '轉入';

  @override
  String get transferNote => '說明：將佣金收益轉入可用餘額後，可用於給玩家充值或申請提現';

  @override
  String get submitFundRequest => '發起資金申請';

  @override
  String get ownerDeposit => '餘額充值';

  @override
  String get ownerWithdraw => '餘額提現';

  @override
  String get marginDeposit2 => '保證金充值';

  @override
  String get amount => '金額';

  @override
  String get remarkOptional => '備註（可選）';

  @override
  String get submitRequest => '提交申請';

  @override
  String get fundRequestNote => '說明：充值/提現申請提交後，需要管理員在後台審核通過後，餘額才會實際變動。';

  @override
  String get fundRequestRecords => '資金申請記錄';

  @override
  String get pending => '待審核';

  @override
  String get approved => '已通過';

  @override
  String get rejected => '已拒絕';

  @override
  String get transactionHistory => '資金流水';

  @override
  String get gameBet => '遊戲下注';

  @override
  String get gameWin => '遊戲獲勝';

  @override
  String get gameRefund => '遊戲退款';

  @override
  String get ownerCommission => '房主佣金';

  @override
  String get platformShare => '平台抽成';

  @override
  String get pleaseEnterAmount => '請輸入金額';

  @override
  String get invalidAmountFormat => '金額格式不正確';

  @override
  String get pleaseSelectType => '請選擇申請類型';

  @override
  String get fundRequestSubmitted => '資金申請已提交，等待審核';

  @override
  String get submitFailed => '提交失敗';

  @override
  String get pleaseEnterTransferAmount => '請輸入轉帳金額';

  @override
  String transferAmountExceeds(String amount) {
    return '轉帳金額不能超過 ¥$amount';
  }

  @override
  String get transferSuccess => '轉帳成功';

  @override
  String get transferFailed => '轉帳失敗';

  @override
  String get pendingPlayerRequests => '待審批的玩家申請';

  @override
  String get noPendingRequests => '暫無待審批的申請';

  @override
  String get unknownUser => '未知用戶';

  @override
  String get confirmApprove => '確認通過';

  @override
  String get confirmReject => '確認拒絕';

  @override
  String get confirmApproveMessage => '確定要通過這個申請嗎？';

  @override
  String get confirmRejectMessage => '確定要拒絕這個申請嗎？';

  @override
  String get approve => '通過';

  @override
  String get reject => '拒絕';

  @override
  String get requestApproved => '申請已通過';

  @override
  String get requestRejected => '申請已拒絕';

  @override
  String get operationFailed => '操作失敗';

  @override
  String get basicInfo => '基本資訊';

  @override
  String get gameSettings => '遊戲設定';

  @override
  String get feeSettings => '費率設定';

  @override
  String get securitySettings => '安全設定';

  @override
  String get platformFeeFixed => '平台費率（固定）';

  @override
  String get ownerFee => '房主費率';

  @override
  String get totalFee => '總費率';

  @override
  String get adjustOwnerFee => '調整房主費率:';

  @override
  String get feeNote => '說明：每局遊戲結算時，從獎池中扣除總費率作為平台和房主收益';

  @override
  String get roomPassword => '房間密碼';

  @override
  String get setPassword => '設定密碼';

  @override
  String get pleaseEnterRoomName => '請輸入房間名稱';

  @override
  String get pleaseEnterPassword => '請輸入密碼';

  @override
  String get createFailed => '創建失敗';

  @override
  String get winnersMustLessThanPlayers => '獲勝人數必須小於最大玩家數';

  @override
  String minPlayersHint(int count) {
    return '每局遊戲的獲勝人數，最少需要 $count 人才能開始遊戲';
  }

  @override
  String get maxPlayersHint => '房間最多容納的玩家數';

  @override
  String get connectionError => '連接錯誤';

  @override
  String get connectionClosed => '連接已斷開，正在重連...';

  @override
  String get joinRoomFailed => '連接房間失敗';

  @override
  String get joinRoomTitle => '加入房間';

  @override
  String get currentBalance => '當前餘額';

  @override
  String get minBalanceRequired => '最低餘額要求';

  @override
  String get insufficientBalanceCannotJoin => '餘額不足，無法加入房間';

  @override
  String get riskWarning => '風險提示';

  @override
  String get riskWarningTip1 => '遊戲有風險，請理性參與';

  @override
  String get riskWarningTip2 => '進入房間後將自動參與每輪遊戲';

  @override
  String get confirmJoin => '確認加入';

  @override
  String get minPlayersRequired => '最少開始人數';

  @override
  String get winnersPerRound => '每輪贏家數';

  @override
  String get maxPlayersAllowed => '最大參與人數';

  @override
  String get betPerRoundLabel => '每輪下注';

  @override
  String get personUnit => '人';
}
