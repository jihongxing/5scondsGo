// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => '5SecondsGo';

  @override
  String get login => 'Login';

  @override
  String get register => 'Register';

  @override
  String get username => 'Username';

  @override
  String get password => 'Password';

  @override
  String get inviteCode => 'Invite Code';

  @override
  String get loginButton => 'Sign In';

  @override
  String get registerButton => 'Sign Up';

  @override
  String get noAccount => 'Don\'t have an account?';

  @override
  String get hasAccount => 'Already have an account?';

  @override
  String get home => 'Home';

  @override
  String get rooms => 'Rooms';

  @override
  String get profile => 'Profile';

  @override
  String get wallet => 'Wallet';

  @override
  String get joinRoom => 'Join Room';

  @override
  String get leaveRoom => 'Leave Room';

  @override
  String get createRoom => 'Create Room';

  @override
  String get roomName => 'Room Name';

  @override
  String get betAmount => 'Bet Amount';

  @override
  String get winnerCount => 'Winner Count';

  @override
  String get maxPlayers => 'Max Players';

  @override
  String get commissionRate => 'Commission Rate';

  @override
  String get phaseWaiting => 'Waiting';

  @override
  String get phaseCountdown => 'Countdown';

  @override
  String get phaseBetting => 'Betting';

  @override
  String get phaseInGame => 'In Game';

  @override
  String get phaseSettlement => 'Settlement';

  @override
  String get phaseReset => 'Reset';

  @override
  String get autoReady => 'Auto Ready';

  @override
  String get ready => 'Ready';

  @override
  String get notReady => 'Not Ready';

  @override
  String get balance => 'Balance';

  @override
  String get frozenBalance => 'Frozen';

  @override
  String get deposit => 'Deposit';

  @override
  String get withdraw => 'Withdraw';

  @override
  String get currentRound => 'Round';

  @override
  String get poolAmount => 'Pool';

  @override
  String get winners => 'Winners';

  @override
  String get prize => 'Prize';

  @override
  String get players => 'Players';

  @override
  String get online => 'Online';

  @override
  String get offline => 'Offline';

  @override
  String get submit => 'Submit';

  @override
  String get cancel => 'Cancel';

  @override
  String get confirm => 'Confirm';

  @override
  String get close => 'Close';

  @override
  String get save => 'Save';

  @override
  String get error => 'Error';

  @override
  String get success => 'Success';

  @override
  String get loading => 'Loading...';

  @override
  String get noData => 'No Data';

  @override
  String get invalidCredentials => 'Invalid username or password';

  @override
  String get usernameExists => 'Username already exists';

  @override
  String get invalidInviteCode => 'Invalid invite code';

  @override
  String get roomFull => 'Room is full';

  @override
  String get insufficientBalance => 'Insufficient balance';

  @override
  String get congratulations => 'Congratulations!';

  @override
  String get youWon => 'You Won!';

  @override
  String get betterLuckNextTime => 'Better Luck Next Time';

  @override
  String get verifyResult => 'Verify Result';

  @override
  String get commitHash => 'Commit Hash';

  @override
  String get revealSeed => 'Reveal Seed';

  @override
  String get roomThemeTitle => 'Room Theme';

  @override
  String get roomThemeOwnerOnly => 'Only room owner can change theme';

  @override
  String get themeClassic => 'Classic';

  @override
  String get themeNeon => 'Neon';

  @override
  String get themeOcean => 'Ocean';

  @override
  String get themeForest => 'Forest';

  @override
  String get themeLuxury => 'Luxury';

  @override
  String get settings => 'Settings';

  @override
  String get language => 'Language';

  @override
  String get logout => 'Logout';

  @override
  String get leaderboard => 'Leaderboard';

  @override
  String get friends => 'Friends';

  @override
  String get spectate => 'Spectate';

  @override
  String get switchToParticipant => 'Join Game';

  @override
  String get betPerRound => 'Bet/Round';

  @override
  String get gameHistory => 'History';

  @override
  String minPlayersToStart(int count) {
    return 'Min $count to start';
  }

  @override
  String get congratsWin => 'Congratulations!';

  @override
  String get wonPrize => 'won';

  @override
  String playerWon(String names) {
    return '$names won';
  }

  @override
  String bettingComplete(int count, String pool) {
    return 'Betting complete: $count joined, Pool ¥$pool';
  }

  @override
  String get noWinner => 'No winner';

  @override
  String wonAmount(String names, String prize) {
    return '$names won ¥$prize';
  }

  @override
  String get gameCancelled => 'Game cancelled';

  @override
  String get notEnoughPlayers => 'Not enough players';

  @override
  String get insufficientBalanceDisqualified =>
      'Insufficient balance, you cannot participate in this round';

  @override
  String get disqualifiedFromRound =>
      'You have been disqualified from this round';

  @override
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers) {
    return 'Round cancelled: need $minRequired players, only $currentPlayers qualified';
  }

  @override
  String get gameRecords => 'Records';

  @override
  String get myProfile => 'Profile';

  @override
  String get owner => 'Owner';

  @override
  String get admin => 'Admin';

  @override
  String get player => 'Player';

  @override
  String get commission => 'Commission';

  @override
  String get noRooms => 'No rooms available';

  @override
  String get contactOwnerToCreate => 'Contact owner to create a room';

  @override
  String get enterRoomPassword => 'Enter room password';

  @override
  String get passwordHint => 'Password';

  @override
  String get availableBalance => 'Available';

  @override
  String get totalBalance => 'Total';

  @override
  String get marginDeposit => 'Margin';

  @override
  String get commissionEarnings => 'Commission';

  @override
  String get myInviteCode => 'My Invite Code';

  @override
  String get inviteCodeCopied => 'Invite code copied';

  @override
  String get selectLanguage => 'Select Language';

  @override
  String get confirmLogout => 'Confirm Logout';

  @override
  String get confirmLogoutMessage => 'Are you sure you want to log out?';

  @override
  String get loadFailed => 'Load failed';

  @override
  String get retry => 'Retry';

  @override
  String get noFriends => 'No friends yet';

  @override
  String get clickToAddFriend => 'Click top right to add friends';

  @override
  String get searchFriends => 'Search friends...';

  @override
  String get addFriend => 'Add Friend';

  @override
  String get userId => 'User ID';

  @override
  String get enterUserId => 'Enter user ID';

  @override
  String get send => 'Send';

  @override
  String get friendRequestSent => 'Friend request sent';

  @override
  String get sendFailed => 'Send failed';

  @override
  String get inRoom => 'In room';

  @override
  String get inviteToRoom => 'Invite to room';

  @override
  String get removeFriend => 'Remove friend';

  @override
  String get confirmRemoveFriend => 'Remove Friend';

  @override
  String confirmRemoveFriendMessage(String name) {
    return 'Are you sure you want to remove $name?';
  }

  @override
  String get remove => 'Remove';

  @override
  String get totalGames => 'Games';

  @override
  String get wins => 'Wins';

  @override
  String get winRate => 'Win Rate';

  @override
  String get totalBet => 'Total Bet';

  @override
  String get totalWon => 'Total Won';

  @override
  String get netProfit => 'Net P/L';

  @override
  String roundNumber(int number) {
    return 'Round $number';
  }

  @override
  String get skipped => 'Skipped';

  @override
  String get bet => 'Bet';

  @override
  String get ownerWallet => 'Owner Wallet';

  @override
  String get marginBalance => 'Margin';

  @override
  String get commissionBalance => 'Commission';

  @override
  String get totalCommission => 'Total Commission';

  @override
  String get canRechargePlayer => 'For player recharge';

  @override
  String get fixedGuarantee => 'Fixed guarantee';

  @override
  String get canTransferToBalance => 'Can transfer to balance';

  @override
  String get transferToBalance => 'Transfer to Balance';

  @override
  String get transferAmount => 'Transfer Amount';

  @override
  String maxTransfer(String amount) {
    return 'Max $amount';
  }

  @override
  String get transfer => 'Transfer';

  @override
  String get transferNote =>
      'Note: Transfer commission to available balance for player recharge or withdrawal';

  @override
  String get submitFundRequest => 'Submit Fund Request';

  @override
  String get ownerDeposit => 'Balance Deposit';

  @override
  String get ownerWithdraw => 'Balance Withdraw';

  @override
  String get marginDeposit2 => 'Margin Deposit';

  @override
  String get amount => 'Amount';

  @override
  String get remarkOptional => 'Remark (Optional)';

  @override
  String get submitRequest => 'Submit Request';

  @override
  String get fundRequestNote =>
      'Note: Deposit/withdrawal requests need admin approval before balance changes';

  @override
  String get fundRequestRecords => 'Fund Request Records';

  @override
  String get pending => 'Pending';

  @override
  String get approved => 'Approved';

  @override
  String get rejected => 'Rejected';

  @override
  String get transactionHistory => 'Transaction History';

  @override
  String get gameBet => 'Game Bet';

  @override
  String get gameWin => 'Game Win';

  @override
  String get gameRefund => 'Game Refund';

  @override
  String get ownerCommission => 'Owner Commission';

  @override
  String get platformShare => 'Platform Share';

  @override
  String get pleaseEnterAmount => 'Please enter amount';

  @override
  String get invalidAmountFormat => 'Invalid amount format';

  @override
  String get pleaseSelectType => 'Please select type';

  @override
  String get fundRequestSubmitted => 'Fund request submitted, pending approval';

  @override
  String get submitFailed => 'Submit failed';

  @override
  String get pleaseEnterTransferAmount => 'Please enter transfer amount';

  @override
  String transferAmountExceeds(String amount) {
    return 'Transfer amount cannot exceed ¥$amount';
  }

  @override
  String get transferSuccess => 'Transfer successful';

  @override
  String get transferFailed => 'Transfer failed';

  @override
  String get pendingPlayerRequests => 'Pending Player Requests';

  @override
  String get noPendingRequests => 'No pending requests';

  @override
  String get unknownUser => 'Unknown User';

  @override
  String get confirmApprove => 'Confirm Approve';

  @override
  String get confirmReject => 'Confirm Reject';

  @override
  String get confirmApproveMessage =>
      'Are you sure you want to approve this request?';

  @override
  String get confirmRejectMessage =>
      'Are you sure you want to reject this request?';

  @override
  String get approve => 'Approve';

  @override
  String get reject => 'Reject';

  @override
  String get requestApproved => 'Request approved';

  @override
  String get requestRejected => 'Request rejected';

  @override
  String get operationFailed => 'Operation failed';

  @override
  String get basicInfo => 'Basic Info';

  @override
  String get gameSettings => 'Game Settings';

  @override
  String get feeSettings => 'Fee Settings';

  @override
  String get securitySettings => 'Security Settings';

  @override
  String get platformFeeFixed => 'Platform Fee (Fixed)';

  @override
  String get ownerFee => 'Owner Fee';

  @override
  String get totalFee => 'Total Fee';

  @override
  String get adjustOwnerFee => 'Adjust Owner Fee:';

  @override
  String get feeNote =>
      'Note: Total fee is deducted from prize pool at settlement';

  @override
  String get roomPassword => 'Room Password';

  @override
  String get setPassword => 'Set Password';

  @override
  String get pleaseEnterRoomName => 'Please enter room name';

  @override
  String get pleaseEnterPassword => 'Please enter password';

  @override
  String get createFailed => 'Create failed';

  @override
  String get winnersMustLessThanPlayers =>
      'Winners must be less than max players';

  @override
  String minPlayersHint(int count) {
    return 'Winners per round, need at least $count players to start';
  }

  @override
  String get maxPlayersHint => 'Maximum players in room';

  @override
  String get connectionError => 'Connection error';

  @override
  String get connectionClosed => 'Connection closed, reconnecting...';

  @override
  String get joinRoomFailed => 'Failed to join room';

  @override
  String get joinRoomTitle => 'Join Room';

  @override
  String get currentBalance => 'Current Balance';

  @override
  String get minBalanceRequired => 'Min Balance Required';

  @override
  String get insufficientBalanceCannotJoin =>
      'Insufficient balance, cannot join room';

  @override
  String get riskWarning => 'Risk Warning';

  @override
  String get riskWarningTip1 =>
      'Gambling involves risk, please participate responsibly';

  @override
  String get riskWarningTip2 =>
      'You will automatically join each round after entering';

  @override
  String get confirmJoin => 'Confirm Join';

  @override
  String get minPlayersRequired => 'Min Players';

  @override
  String get winnersPerRound => 'Winners/Round';

  @override
  String get maxPlayersAllowed => 'Max Players';

  @override
  String get betPerRoundLabel => 'Bet/Round';

  @override
  String get personUnit => '';
}
