// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Korean (`ko`).
class AppLocalizationsKo extends AppLocalizations {
  AppLocalizationsKo([String locale = 'ko']) : super(locale);

  @override
  String get appTitle => '5초 추첨';

  @override
  String get login => '로그인';

  @override
  String get register => '회원가입';

  @override
  String get username => '사용자명';

  @override
  String get password => '비밀번호';

  @override
  String get inviteCode => '초대 코드';

  @override
  String get loginButton => '로그인';

  @override
  String get registerButton => '회원가입';

  @override
  String get noAccount => '계정이 없으신가요?';

  @override
  String get hasAccount => '이미 계정이 있으신가요?';

  @override
  String get home => '홈';

  @override
  String get rooms => '방';

  @override
  String get profile => '프로필';

  @override
  String get wallet => '지갑';

  @override
  String get joinRoom => '방 참가';

  @override
  String get leaveRoom => '방 나가기';

  @override
  String get createRoom => '방 만들기';

  @override
  String get roomName => '방 이름';

  @override
  String get betAmount => '베팅 금액';

  @override
  String get winnerCount => '당첨자 수';

  @override
  String get maxPlayers => '최대 인원';

  @override
  String get commissionRate => '수수료율';

  @override
  String get phaseWaiting => '대기 중';

  @override
  String get phaseCountdown => '카운트다운';

  @override
  String get phaseBetting => '베팅 중';

  @override
  String get phaseInGame => '게임 중';

  @override
  String get phaseSettlement => '정산 중';

  @override
  String get phaseReset => '리셋 중';

  @override
  String get autoReady => '자동 준비';

  @override
  String get ready => '준비 완료';

  @override
  String get notReady => '준비 안됨';

  @override
  String get balance => '잔액';

  @override
  String get frozenBalance => '동결';

  @override
  String get deposit => '입금';

  @override
  String get withdraw => '출금';

  @override
  String get currentRound => '라운드';

  @override
  String get poolAmount => '상금 풀';

  @override
  String get winners => '당첨자';

  @override
  String get prize => '상금';

  @override
  String get players => '플레이어';

  @override
  String get online => '온라인';

  @override
  String get offline => '오프라인';

  @override
  String get submit => '제출';

  @override
  String get cancel => '취소';

  @override
  String get confirm => '확인';

  @override
  String get close => '닫기';

  @override
  String get save => '저장';

  @override
  String get error => '오류';

  @override
  String get success => '성공';

  @override
  String get loading => '로딩 중...';

  @override
  String get noData => '데이터 없음';

  @override
  String get invalidCredentials => '사용자명 또는 비밀번호가 올바르지 않습니다';

  @override
  String get usernameExists => '사용자명이 이미 존재합니다';

  @override
  String get invalidInviteCode => '초대 코드가 유효하지 않습니다';

  @override
  String get roomFull => '방이 가득 찼습니다';

  @override
  String get insufficientBalance => '잔액 부족';

  @override
  String get congratulations => '축하합니다!';

  @override
  String get youWon => '당첨되었습니다!';

  @override
  String get betterLuckNextTime => '다음에 행운을 빕니다';

  @override
  String get verifyResult => '결과 검증';

  @override
  String get commitHash => '커밋 해시';

  @override
  String get revealSeed => '시드 값';

  @override
  String get roomThemeTitle => '방 테마';

  @override
  String get roomThemeOwnerOnly => '테마는 방장만 변경할 수 있습니다';

  @override
  String get themeClassic => '클래식';

  @override
  String get themeNeon => '네온';

  @override
  String get themeOcean => '오션';

  @override
  String get themeForest => '포레스트';

  @override
  String get themeLuxury => '럭셔리';

  @override
  String get settings => '설정';

  @override
  String get language => '언어';

  @override
  String get logout => '로그아웃';

  @override
  String get leaderboard => '랭킹';

  @override
  String get friends => '친구';

  @override
  String get spectate => '관전';

  @override
  String get switchToParticipant => '참가하기';

  @override
  String get betPerRound => '라운드당 베팅';

  @override
  String get gameHistory => '게임 기록';

  @override
  String minPlayersToStart(int count) {
    return '최소 $count명 시작';
  }

  @override
  String get congratsWin => '축하합니다!';

  @override
  String get wonPrize => '획득';

  @override
  String playerWon(String names) {
    return '$names 당첨';
  }

  @override
  String bettingComplete(int count, String pool) {
    return '베팅 완료: $count명 참가, 상금 풀 ¥$pool';
  }

  @override
  String get noWinner => '당첨자 없음';

  @override
  String wonAmount(String names, String prize) {
    return '$names이(가) ¥$prize 획득';
  }

  @override
  String get gameCancelled => '게임 취소';

  @override
  String get notEnoughPlayers => '인원 부족';

  @override
  String get insufficientBalanceDisqualified => '잔액 부족으로 이번 라운드에 참가할 수 없습니다';

  @override
  String get disqualifiedFromRound => '이번 라운드에서 실격되었습니다';

  @override
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers) {
    return '라운드 취소: $minRequired명 필요, $currentPlayers명만 자격 있음';
  }

  @override
  String get gameRecords => '게임 기록';

  @override
  String get myProfile => '내 정보';

  @override
  String get owner => '방장';

  @override
  String get admin => '관리자';

  @override
  String get player => '플레이어';

  @override
  String get commission => '수수료';

  @override
  String get noRooms => '방이 없습니다';

  @override
  String get contactOwnerToCreate => '방장에게 연락하여 방 생성';

  @override
  String get enterRoomPassword => '방 비밀번호 입력';

  @override
  String get passwordHint => '비밀번호';

  @override
  String get availableBalance => '사용 가능 잔액';

  @override
  String get totalBalance => '총 잔액';

  @override
  String get marginDeposit => '보증금';

  @override
  String get commissionEarnings => '수수료 수익';

  @override
  String get myInviteCode => '내 초대 코드';

  @override
  String get inviteCodeCopied => '초대 코드가 복사되었습니다';

  @override
  String get selectLanguage => '언어 선택';

  @override
  String get confirmLogout => '로그아웃 확인';

  @override
  String get confirmLogoutMessage => '로그아웃 하시겠습니까?';

  @override
  String get loadFailed => '로드 실패';

  @override
  String get retry => '재시도';

  @override
  String get noFriends => '친구가 없습니다';

  @override
  String get clickToAddFriend => '오른쪽 상단을 눌러 친구 추가';

  @override
  String get searchFriends => '친구 검색...';

  @override
  String get addFriend => '친구 추가';

  @override
  String get userId => '사용자 ID';

  @override
  String get enterUserId => '사용자 ID 입력';

  @override
  String get send => '보내기';

  @override
  String get friendRequestSent => '친구 요청을 보냈습니다';

  @override
  String get sendFailed => '전송 실패';

  @override
  String get inRoom => '방에 있음';

  @override
  String get inviteToRoom => '방에 초대';

  @override
  String get removeFriend => '친구 삭제';

  @override
  String get confirmRemoveFriend => '친구 삭제';

  @override
  String confirmRemoveFriendMessage(String name) {
    return '$name을(를) 삭제하시겠습니까?';
  }

  @override
  String get remove => '삭제';

  @override
  String get totalGames => '총 게임';

  @override
  String get wins => '승리';

  @override
  String get winRate => '승률';

  @override
  String get totalBet => '총 베팅';

  @override
  String get totalWon => '총 획득';

  @override
  String get netProfit => '순손익';

  @override
  String roundNumber(int number) {
    return '$number라운드';
  }

  @override
  String get skipped => '건너뜀';

  @override
  String get bet => '베팅';

  @override
  String get ownerWallet => '방장 지갑';

  @override
  String get marginBalance => '보증금';

  @override
  String get commissionBalance => '수수료 수익';

  @override
  String get totalCommission => '누적 수수료';

  @override
  String get canRechargePlayer => '플레이어 충전용';

  @override
  String get fixedGuarantee => '고정 보증';

  @override
  String get canTransferToBalance => '잔액으로 이체 가능';

  @override
  String get transferToBalance => '수익을 잔액으로 이체';

  @override
  String get transferAmount => '이체 금액';

  @override
  String maxTransfer(String amount) {
    return '최대 $amount';
  }

  @override
  String get transfer => '이체';

  @override
  String get transferNote =>
      '설명: 수수료 수익을 사용 가능 잔액으로 이체하면 플레이어 충전이나 출금 신청에 사용할 수 있습니다';

  @override
  String get submitFundRequest => '자금 신청 제출';

  @override
  String get ownerDeposit => '잔액 입금';

  @override
  String get ownerWithdraw => '잔액 출금';

  @override
  String get marginDeposit2 => '보증금 입금';

  @override
  String get amount => '금액';

  @override
  String get remarkOptional => '비고 (선택)';

  @override
  String get submitRequest => '신청 제출';

  @override
  String get fundRequestNote => '설명: 입금/출금 신청은 관리자 승인 후 잔액이 변경됩니다';

  @override
  String get fundRequestRecords => '자금 신청 기록';

  @override
  String get pending => '심사 중';

  @override
  String get approved => '승인됨';

  @override
  String get rejected => '거절됨';

  @override
  String get transactionHistory => '거래 내역';

  @override
  String get gameBet => '게임 베팅';

  @override
  String get gameWin => '게임 승리';

  @override
  String get gameRefund => '게임 환불';

  @override
  String get ownerCommission => '방장 수수료';

  @override
  String get platformShare => '플랫폼 수수료';

  @override
  String get pleaseEnterAmount => '금액을 입력하세요';

  @override
  String get invalidAmountFormat => '금액 형식이 올바르지 않습니다';

  @override
  String get pleaseSelectType => '신청 유형을 선택하세요';

  @override
  String get fundRequestSubmitted => '자금 신청이 제출되었습니다, 승인 대기 중';

  @override
  String get submitFailed => '제출 실패';

  @override
  String get pleaseEnterTransferAmount => '이체 금액을 입력하세요';

  @override
  String transferAmountExceeds(String amount) {
    return '이체 금액은 ¥$amount을 초과할 수 없습니다';
  }

  @override
  String get transferSuccess => '이체 성공';

  @override
  String get transferFailed => '이체 실패';

  @override
  String get pendingPlayerRequests => '승인 대기 중인 플레이어 신청';

  @override
  String get noPendingRequests => '승인 대기 중인 신청이 없습니다';

  @override
  String get unknownUser => '알 수 없는 사용자';

  @override
  String get confirmApprove => '승인 확인';

  @override
  String get confirmReject => '거절 확인';

  @override
  String get confirmApproveMessage => '이 신청을 승인하시겠습니까?';

  @override
  String get confirmRejectMessage => '이 신청을 거절하시겠습니까?';

  @override
  String get approve => '승인';

  @override
  String get reject => '거절';

  @override
  String get requestApproved => '신청이 승인되었습니다';

  @override
  String get requestRejected => '신청이 거절되었습니다';

  @override
  String get operationFailed => '작업 실패';

  @override
  String get basicInfo => '기본 정보';

  @override
  String get gameSettings => '게임 설정';

  @override
  String get feeSettings => '수수료 설정';

  @override
  String get securitySettings => '보안 설정';

  @override
  String get platformFeeFixed => '플랫폼 수수료 (고정)';

  @override
  String get ownerFee => '방장 수수료';

  @override
  String get totalFee => '총 수수료';

  @override
  String get adjustOwnerFee => '방장 수수료 조정:';

  @override
  String get feeNote => '설명: 각 게임 정산 시 상금 풀에서 총 수수료가 차감됩니다';

  @override
  String get roomPassword => '방 비밀번호';

  @override
  String get setPassword => '비밀번호 설정';

  @override
  String get pleaseEnterRoomName => '방 이름을 입력하세요';

  @override
  String get pleaseEnterPassword => '비밀번호를 입력하세요';

  @override
  String get createFailed => '생성 실패';

  @override
  String get winnersMustLessThanPlayers => '당첨자 수는 최대 플레이어 수보다 적어야 합니다';

  @override
  String minPlayersHint(int count) {
    return '라운드당 당첨자 수, 최소 $count명으로 게임 시작';
  }

  @override
  String get maxPlayersHint => '방 최대 수용 인원';

  @override
  String get connectionError => '연결 오류';

  @override
  String get connectionClosed => '연결이 끊어졌습니다, 재연결 중...';

  @override
  String get joinRoomFailed => '방 연결 실패';

  @override
  String get joinRoomTitle => '방 참가';

  @override
  String get currentBalance => '현재 잔액';

  @override
  String get minBalanceRequired => '최소 잔액 요건';

  @override
  String get insufficientBalanceCannotJoin => '잔액 부족으로 방에 참가할 수 없습니다';

  @override
  String get riskWarning => '위험 경고';

  @override
  String get riskWarningTip1 => '게임에는 위험이 있습니다, 책임감 있게 참여하세요';

  @override
  String get riskWarningTip2 => '입장 후 각 라운드에 자동으로 참가합니다';

  @override
  String get confirmJoin => '참가 확인';

  @override
  String get minPlayersRequired => '최소 시작 인원';

  @override
  String get winnersPerRound => '라운드 당첨자 수';

  @override
  String get maxPlayersAllowed => '최대 참가 인원';

  @override
  String get betPerRoundLabel => '라운드 베팅';

  @override
  String get personUnit => '명';
}
