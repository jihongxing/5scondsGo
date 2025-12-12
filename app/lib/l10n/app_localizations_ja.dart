// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Japanese (`ja`).
class AppLocalizationsJa extends AppLocalizations {
  AppLocalizationsJa([String locale = 'ja']) : super(locale);

  @override
  String get appTitle => '5秒抽選';

  @override
  String get login => 'ログイン';

  @override
  String get register => '登録';

  @override
  String get username => 'ユーザー名';

  @override
  String get password => 'パスワード';

  @override
  String get inviteCode => '招待コード';

  @override
  String get loginButton => 'ログイン';

  @override
  String get registerButton => '登録';

  @override
  String get noAccount => 'アカウントをお持ちでない方';

  @override
  String get hasAccount => 'すでにアカウントをお持ちの方';

  @override
  String get home => 'ホーム';

  @override
  String get rooms => 'ルーム';

  @override
  String get profile => 'プロフィール';

  @override
  String get wallet => 'ウォレット';

  @override
  String get joinRoom => 'ルームに参加';

  @override
  String get leaveRoom => 'ルームを退出';

  @override
  String get createRoom => 'ルームを作成';

  @override
  String get roomName => 'ルーム名';

  @override
  String get betAmount => 'ベット額';

  @override
  String get winnerCount => '当選者数';

  @override
  String get maxPlayers => '最大人数';

  @override
  String get commissionRate => '手数料率';

  @override
  String get phaseWaiting => '待機中';

  @override
  String get phaseCountdown => 'カウントダウン';

  @override
  String get phaseBetting => 'ベット中';

  @override
  String get phaseInGame => 'ゲーム中';

  @override
  String get phaseSettlement => '精算中';

  @override
  String get phaseReset => 'リセット中';

  @override
  String get autoReady => '自動準備';

  @override
  String get ready => '準備完了';

  @override
  String get notReady => '未準備';

  @override
  String get balance => '残高';

  @override
  String get frozenBalance => '凍結';

  @override
  String get deposit => '入金';

  @override
  String get withdraw => '出金';

  @override
  String get currentRound => 'ラウンド';

  @override
  String get poolAmount => '賞金プール';

  @override
  String get winners => '当選者';

  @override
  String get prize => '賞金';

  @override
  String get players => 'プレイヤー';

  @override
  String get online => 'オンライン';

  @override
  String get offline => 'オフライン';

  @override
  String get submit => '送信';

  @override
  String get cancel => 'キャンセル';

  @override
  String get confirm => '確認';

  @override
  String get close => '閉じる';

  @override
  String get save => '保存';

  @override
  String get error => 'エラー';

  @override
  String get success => '成功';

  @override
  String get loading => '読み込み中...';

  @override
  String get noData => 'データなし';

  @override
  String get invalidCredentials => 'ユーザー名またはパスワードが正しくありません';

  @override
  String get usernameExists => 'ユーザー名は既に存在します';

  @override
  String get invalidInviteCode => '招待コードが無効です';

  @override
  String get roomFull => 'ルームが満員です';

  @override
  String get insufficientBalance => '残高不足';

  @override
  String get congratulations => 'おめでとうございます！';

  @override
  String get youWon => '当選しました！';

  @override
  String get betterLuckNextTime => '次回頑張ってください';

  @override
  String get verifyResult => '結果を検証';

  @override
  String get commitHash => 'コミットハッシュ';

  @override
  String get revealSeed => 'シード値';

  @override
  String get roomThemeTitle => 'ルームテーマ';

  @override
  String get roomThemeOwnerOnly => 'テーマはルームオーナーのみ変更可能';

  @override
  String get themeClassic => 'クラシック';

  @override
  String get themeNeon => 'ネオン';

  @override
  String get themeOcean => 'オーシャン';

  @override
  String get themeForest => 'フォレスト';

  @override
  String get themeLuxury => 'ラグジュアリー';

  @override
  String get settings => '設定';

  @override
  String get language => '言語';

  @override
  String get logout => 'ログアウト';

  @override
  String get leaderboard => 'ランキング';

  @override
  String get friends => 'フレンド';

  @override
  String get spectate => '観戦';

  @override
  String get switchToParticipant => '参加する';

  @override
  String get betPerRound => '1回のベット';

  @override
  String get gameHistory => 'ゲーム履歴';

  @override
  String minPlayersToStart(int count) {
    return '最低$count人で開始';
  }

  @override
  String get congratsWin => 'おめでとう！';

  @override
  String get wonPrize => '獲得';

  @override
  String playerWon(String names) {
    return '$names が当選';
  }

  @override
  String bettingComplete(int count, String pool) {
    return 'ベット完了：$count人参加、賞金プール ¥$pool';
  }

  @override
  String get noWinner => '当選者なし';

  @override
  String wonAmount(String names, String prize) {
    return '$names が ¥$prize を獲得';
  }

  @override
  String get gameCancelled => 'ゲームキャンセル';

  @override
  String get notEnoughPlayers => '人数不足';

  @override
  String get insufficientBalanceDisqualified => '残高不足のため、このラウンドに参加できません';

  @override
  String get disqualifiedFromRound => 'このラウンドから失格となりました';

  @override
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers) {
    return 'ラウンドキャンセル：$minRequired人必要、$currentPlayers人のみ資格あり';
  }

  @override
  String get gameRecords => 'ゲーム履歴';

  @override
  String get myProfile => 'マイページ';

  @override
  String get owner => 'オーナー';

  @override
  String get admin => '管理者';

  @override
  String get player => 'プレイヤー';

  @override
  String get commission => '手数料';

  @override
  String get noRooms => 'ルームがありません';

  @override
  String get contactOwnerToCreate => 'オーナーに連絡してルームを作成';

  @override
  String get enterRoomPassword => 'ルームパスワードを入力';

  @override
  String get passwordHint => 'パスワード';

  @override
  String get availableBalance => '利用可能残高';

  @override
  String get totalBalance => '合計残高';

  @override
  String get marginDeposit => '保証金';

  @override
  String get commissionEarnings => '手数料収益';

  @override
  String get myInviteCode => '招待コード';

  @override
  String get inviteCodeCopied => '招待コードをコピーしました';

  @override
  String get selectLanguage => '言語を選択';

  @override
  String get confirmLogout => 'ログアウト確認';

  @override
  String get confirmLogoutMessage => 'ログアウトしますか？';

  @override
  String get loadFailed => '読み込み失敗';

  @override
  String get retry => '再試行';

  @override
  String get noFriends => 'フレンドがいません';

  @override
  String get clickToAddFriend => '右上をタップしてフレンドを追加';

  @override
  String get searchFriends => 'フレンドを検索...';

  @override
  String get addFriend => 'フレンド追加';

  @override
  String get userId => 'ユーザーID';

  @override
  String get enterUserId => 'ユーザーIDを入力';

  @override
  String get send => '送信';

  @override
  String get friendRequestSent => 'フレンドリクエストを送信しました';

  @override
  String get sendFailed => '送信失敗';

  @override
  String get inRoom => 'ルーム内';

  @override
  String get inviteToRoom => 'ルームに招待';

  @override
  String get removeFriend => 'フレンド削除';

  @override
  String get confirmRemoveFriend => 'フレンド削除';

  @override
  String confirmRemoveFriendMessage(String name) {
    return '$nameを削除しますか？';
  }

  @override
  String get remove => '削除';

  @override
  String get totalGames => '総ゲーム数';

  @override
  String get wins => '勝利';

  @override
  String get winRate => '勝率';

  @override
  String get totalBet => '総ベット';

  @override
  String get totalWon => '総獲得';

  @override
  String get netProfit => '純損益';

  @override
  String roundNumber(int number) {
    return '第$numberラウンド';
  }

  @override
  String get skipped => 'スキップ';

  @override
  String get bet => 'ベット';

  @override
  String get ownerWallet => 'オーナーウォレット';

  @override
  String get marginBalance => '保証金';

  @override
  String get commissionBalance => '手数料収益';

  @override
  String get totalCommission => '累計手数料';

  @override
  String get canRechargePlayer => 'プレイヤーへの入金用';

  @override
  String get fixedGuarantee => '固定保証';

  @override
  String get canTransferToBalance => '残高に転送可能';

  @override
  String get transferToBalance => '収益を残高に転送';

  @override
  String get transferAmount => '転送金額';

  @override
  String maxTransfer(String amount) {
    return '最大 $amount';
  }

  @override
  String get transfer => '転送';

  @override
  String get transferNote => '説明：手数料収益を利用可能残高に転送すると、プレイヤーへの入金や出金申請に使用できます';

  @override
  String get submitFundRequest => '資金申請を提出';

  @override
  String get ownerDeposit => '残高入金';

  @override
  String get ownerWithdraw => '残高出金';

  @override
  String get marginDeposit2 => '保証金入金';

  @override
  String get amount => '金額';

  @override
  String get remarkOptional => '備考（任意）';

  @override
  String get submitRequest => '申請を提出';

  @override
  String get fundRequestNote => '説明：入金/出金申請は管理者の承認後に残高が変動します';

  @override
  String get fundRequestRecords => '資金申請履歴';

  @override
  String get pending => '審査中';

  @override
  String get approved => '承認済み';

  @override
  String get rejected => '却下';

  @override
  String get transactionHistory => '取引履歴';

  @override
  String get gameBet => 'ゲームベット';

  @override
  String get gameWin => 'ゲーム勝利';

  @override
  String get gameRefund => 'ゲーム返金';

  @override
  String get ownerCommission => 'オーナー手数料';

  @override
  String get platformShare => 'プラットフォーム手数料';

  @override
  String get pleaseEnterAmount => '金額を入力してください';

  @override
  String get invalidAmountFormat => '金額の形式が正しくありません';

  @override
  String get pleaseSelectType => '申請タイプを選択してください';

  @override
  String get fundRequestSubmitted => '資金申請が提出されました、承認待ち';

  @override
  String get submitFailed => '提出失敗';

  @override
  String get pleaseEnterTransferAmount => '転送金額を入力してください';

  @override
  String transferAmountExceeds(String amount) {
    return '転送金額は ¥$amount を超えることはできません';
  }

  @override
  String get transferSuccess => '転送成功';

  @override
  String get transferFailed => '転送失敗';

  @override
  String get pendingPlayerRequests => '承認待ちのプレイヤー申請';

  @override
  String get noPendingRequests => '承認待ちの申請はありません';

  @override
  String get unknownUser => '不明なユーザー';

  @override
  String get confirmApprove => '承認確認';

  @override
  String get confirmReject => '却下確認';

  @override
  String get confirmApproveMessage => 'この申請を承認しますか？';

  @override
  String get confirmRejectMessage => 'この申請を却下しますか？';

  @override
  String get approve => '承認';

  @override
  String get reject => '却下';

  @override
  String get requestApproved => '申請が承認されました';

  @override
  String get requestRejected => '申請が却下されました';

  @override
  String get operationFailed => '操作失敗';

  @override
  String get basicInfo => '基本情報';

  @override
  String get gameSettings => 'ゲーム設定';

  @override
  String get feeSettings => '手数料設定';

  @override
  String get securitySettings => 'セキュリティ設定';

  @override
  String get platformFeeFixed => 'プラットフォーム手数料（固定）';

  @override
  String get ownerFee => 'オーナー手数料';

  @override
  String get totalFee => '合計手数料';

  @override
  String get adjustOwnerFee => 'オーナー手数料を調整:';

  @override
  String get feeNote => '説明：各ゲーム終了時に賞金プールから合計手数料が差し引かれます';

  @override
  String get roomPassword => 'ルームパスワード';

  @override
  String get setPassword => 'パスワードを設定';

  @override
  String get pleaseEnterRoomName => 'ルーム名を入力してください';

  @override
  String get pleaseEnterPassword => 'パスワードを入力してください';

  @override
  String get createFailed => '作成失敗';

  @override
  String get winnersMustLessThanPlayers => '当選者数は最大プレイヤー数より少なくする必要があります';

  @override
  String minPlayersHint(int count) {
    return '1ラウンドの当選者数、最低$count人でゲーム開始';
  }

  @override
  String get maxPlayersHint => 'ルームの最大収容人数';

  @override
  String get connectionError => '接続エラー';

  @override
  String get connectionClosed => '接続が切断されました、再接続中...';

  @override
  String get joinRoomFailed => 'ルームへの接続に失敗しました';

  @override
  String get joinRoomTitle => 'ルームに参加';

  @override
  String get currentBalance => '現在の残高';

  @override
  String get minBalanceRequired => '最低残高要件';

  @override
  String get insufficientBalanceCannotJoin => '残高不足のため、ルームに参加できません';

  @override
  String get riskWarning => 'リスク警告';

  @override
  String get riskWarningTip1 => 'ゲームにはリスクがあります、責任を持って参加してください';

  @override
  String get riskWarningTip2 => '入室後、各ラウンドに自動的に参加します';

  @override
  String get confirmJoin => '参加を確認';

  @override
  String get minPlayersRequired => '最少開始人数';

  @override
  String get winnersPerRound => 'ラウンド当選者数';

  @override
  String get maxPlayersAllowed => '最大参加人数';

  @override
  String get betPerRoundLabel => 'ラウンドベット';

  @override
  String get personUnit => '人';
}
