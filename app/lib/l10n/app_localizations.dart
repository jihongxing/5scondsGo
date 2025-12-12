import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_ja.dart';
import 'app_localizations_ko.dart';
import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('ja'),
    Locale('ko'),
    Locale('zh'),
    Locale('zh', 'TW')
  ];

  /// App title
  ///
  /// In en, this message translates to:
  /// **'5SecondsGo'**
  String get appTitle;

  /// No description provided for @login.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get login;

  /// No description provided for @register.
  ///
  /// In en, this message translates to:
  /// **'Register'**
  String get register;

  /// No description provided for @username.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get username;

  /// No description provided for @password.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get password;

  /// No description provided for @inviteCode.
  ///
  /// In en, this message translates to:
  /// **'Invite Code'**
  String get inviteCode;

  /// No description provided for @loginButton.
  ///
  /// In en, this message translates to:
  /// **'Sign In'**
  String get loginButton;

  /// No description provided for @registerButton.
  ///
  /// In en, this message translates to:
  /// **'Sign Up'**
  String get registerButton;

  /// No description provided for @noAccount.
  ///
  /// In en, this message translates to:
  /// **'Don\'t have an account?'**
  String get noAccount;

  /// No description provided for @hasAccount.
  ///
  /// In en, this message translates to:
  /// **'Already have an account?'**
  String get hasAccount;

  /// No description provided for @home.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get home;

  /// No description provided for @rooms.
  ///
  /// In en, this message translates to:
  /// **'Rooms'**
  String get rooms;

  /// No description provided for @profile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// No description provided for @wallet.
  ///
  /// In en, this message translates to:
  /// **'Wallet'**
  String get wallet;

  /// No description provided for @joinRoom.
  ///
  /// In en, this message translates to:
  /// **'Join Room'**
  String get joinRoom;

  /// No description provided for @leaveRoom.
  ///
  /// In en, this message translates to:
  /// **'Leave Room'**
  String get leaveRoom;

  /// No description provided for @createRoom.
  ///
  /// In en, this message translates to:
  /// **'Create Room'**
  String get createRoom;

  /// No description provided for @roomName.
  ///
  /// In en, this message translates to:
  /// **'Room Name'**
  String get roomName;

  /// No description provided for @betAmount.
  ///
  /// In en, this message translates to:
  /// **'Bet Amount'**
  String get betAmount;

  /// No description provided for @winnerCount.
  ///
  /// In en, this message translates to:
  /// **'Winner Count'**
  String get winnerCount;

  /// No description provided for @maxPlayers.
  ///
  /// In en, this message translates to:
  /// **'Max Players'**
  String get maxPlayers;

  /// No description provided for @commissionRate.
  ///
  /// In en, this message translates to:
  /// **'Commission Rate'**
  String get commissionRate;

  /// No description provided for @phaseWaiting.
  ///
  /// In en, this message translates to:
  /// **'Waiting'**
  String get phaseWaiting;

  /// No description provided for @phaseCountdown.
  ///
  /// In en, this message translates to:
  /// **'Countdown'**
  String get phaseCountdown;

  /// No description provided for @phaseBetting.
  ///
  /// In en, this message translates to:
  /// **'Betting'**
  String get phaseBetting;

  /// No description provided for @phaseInGame.
  ///
  /// In en, this message translates to:
  /// **'In Game'**
  String get phaseInGame;

  /// No description provided for @phaseSettlement.
  ///
  /// In en, this message translates to:
  /// **'Settlement'**
  String get phaseSettlement;

  /// No description provided for @phaseReset.
  ///
  /// In en, this message translates to:
  /// **'Reset'**
  String get phaseReset;

  /// No description provided for @autoReady.
  ///
  /// In en, this message translates to:
  /// **'Auto Ready'**
  String get autoReady;

  /// No description provided for @ready.
  ///
  /// In en, this message translates to:
  /// **'Ready'**
  String get ready;

  /// No description provided for @notReady.
  ///
  /// In en, this message translates to:
  /// **'Not Ready'**
  String get notReady;

  /// No description provided for @balance.
  ///
  /// In en, this message translates to:
  /// **'Balance'**
  String get balance;

  /// No description provided for @frozenBalance.
  ///
  /// In en, this message translates to:
  /// **'Frozen'**
  String get frozenBalance;

  /// No description provided for @deposit.
  ///
  /// In en, this message translates to:
  /// **'Deposit'**
  String get deposit;

  /// No description provided for @withdraw.
  ///
  /// In en, this message translates to:
  /// **'Withdraw'**
  String get withdraw;

  /// No description provided for @currentRound.
  ///
  /// In en, this message translates to:
  /// **'Round'**
  String get currentRound;

  /// No description provided for @poolAmount.
  ///
  /// In en, this message translates to:
  /// **'Pool'**
  String get poolAmount;

  /// No description provided for @winners.
  ///
  /// In en, this message translates to:
  /// **'Winners'**
  String get winners;

  /// No description provided for @prize.
  ///
  /// In en, this message translates to:
  /// **'Prize'**
  String get prize;

  /// No description provided for @players.
  ///
  /// In en, this message translates to:
  /// **'Players'**
  String get players;

  /// No description provided for @online.
  ///
  /// In en, this message translates to:
  /// **'Online'**
  String get online;

  /// No description provided for @offline.
  ///
  /// In en, this message translates to:
  /// **'Offline'**
  String get offline;

  /// No description provided for @submit.
  ///
  /// In en, this message translates to:
  /// **'Submit'**
  String get submit;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @confirm.
  ///
  /// In en, this message translates to:
  /// **'Confirm'**
  String get confirm;

  /// No description provided for @close.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get close;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @error.
  ///
  /// In en, this message translates to:
  /// **'Error'**
  String get error;

  /// No description provided for @success.
  ///
  /// In en, this message translates to:
  /// **'Success'**
  String get success;

  /// No description provided for @loading.
  ///
  /// In en, this message translates to:
  /// **'Loading...'**
  String get loading;

  /// No description provided for @noData.
  ///
  /// In en, this message translates to:
  /// **'No Data'**
  String get noData;

  /// No description provided for @invalidCredentials.
  ///
  /// In en, this message translates to:
  /// **'Invalid username or password'**
  String get invalidCredentials;

  /// No description provided for @usernameExists.
  ///
  /// In en, this message translates to:
  /// **'Username already exists'**
  String get usernameExists;

  /// No description provided for @invalidInviteCode.
  ///
  /// In en, this message translates to:
  /// **'Invalid invite code'**
  String get invalidInviteCode;

  /// No description provided for @roomFull.
  ///
  /// In en, this message translates to:
  /// **'Room is full'**
  String get roomFull;

  /// No description provided for @insufficientBalance.
  ///
  /// In en, this message translates to:
  /// **'Insufficient balance'**
  String get insufficientBalance;

  /// No description provided for @congratulations.
  ///
  /// In en, this message translates to:
  /// **'Congratulations!'**
  String get congratulations;

  /// No description provided for @youWon.
  ///
  /// In en, this message translates to:
  /// **'You Won!'**
  String get youWon;

  /// No description provided for @betterLuckNextTime.
  ///
  /// In en, this message translates to:
  /// **'Better Luck Next Time'**
  String get betterLuckNextTime;

  /// No description provided for @verifyResult.
  ///
  /// In en, this message translates to:
  /// **'Verify Result'**
  String get verifyResult;

  /// No description provided for @commitHash.
  ///
  /// In en, this message translates to:
  /// **'Commit Hash'**
  String get commitHash;

  /// No description provided for @revealSeed.
  ///
  /// In en, this message translates to:
  /// **'Reveal Seed'**
  String get revealSeed;

  /// No description provided for @roomThemeTitle.
  ///
  /// In en, this message translates to:
  /// **'Room Theme'**
  String get roomThemeTitle;

  /// No description provided for @roomThemeOwnerOnly.
  ///
  /// In en, this message translates to:
  /// **'Only room owner can change theme'**
  String get roomThemeOwnerOnly;

  /// No description provided for @themeClassic.
  ///
  /// In en, this message translates to:
  /// **'Classic'**
  String get themeClassic;

  /// No description provided for @themeNeon.
  ///
  /// In en, this message translates to:
  /// **'Neon'**
  String get themeNeon;

  /// No description provided for @themeOcean.
  ///
  /// In en, this message translates to:
  /// **'Ocean'**
  String get themeOcean;

  /// No description provided for @themeForest.
  ///
  /// In en, this message translates to:
  /// **'Forest'**
  String get themeForest;

  /// No description provided for @themeLuxury.
  ///
  /// In en, this message translates to:
  /// **'Luxury'**
  String get themeLuxury;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @logout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// No description provided for @leaderboard.
  ///
  /// In en, this message translates to:
  /// **'Leaderboard'**
  String get leaderboard;

  /// No description provided for @friends.
  ///
  /// In en, this message translates to:
  /// **'Friends'**
  String get friends;

  /// No description provided for @spectate.
  ///
  /// In en, this message translates to:
  /// **'Spectate'**
  String get spectate;

  /// No description provided for @switchToParticipant.
  ///
  /// In en, this message translates to:
  /// **'Join Game'**
  String get switchToParticipant;

  /// No description provided for @betPerRound.
  ///
  /// In en, this message translates to:
  /// **'Bet/Round'**
  String get betPerRound;

  /// No description provided for @gameHistory.
  ///
  /// In en, this message translates to:
  /// **'History'**
  String get gameHistory;

  /// No description provided for @minPlayersToStart.
  ///
  /// In en, this message translates to:
  /// **'Min {count} to start'**
  String minPlayersToStart(int count);

  /// No description provided for @congratsWin.
  ///
  /// In en, this message translates to:
  /// **'Congratulations!'**
  String get congratsWin;

  /// No description provided for @wonPrize.
  ///
  /// In en, this message translates to:
  /// **'won'**
  String get wonPrize;

  /// No description provided for @playerWon.
  ///
  /// In en, this message translates to:
  /// **'{names} won'**
  String playerWon(String names);

  /// No description provided for @bettingComplete.
  ///
  /// In en, this message translates to:
  /// **'Betting complete: {count} joined, Pool ¥{pool}'**
  String bettingComplete(int count, String pool);

  /// No description provided for @noWinner.
  ///
  /// In en, this message translates to:
  /// **'No winner'**
  String get noWinner;

  /// No description provided for @wonAmount.
  ///
  /// In en, this message translates to:
  /// **'{names} won ¥{prize}'**
  String wonAmount(String names, String prize);

  /// No description provided for @gameCancelled.
  ///
  /// In en, this message translates to:
  /// **'Game cancelled'**
  String get gameCancelled;

  /// No description provided for @notEnoughPlayers.
  ///
  /// In en, this message translates to:
  /// **'Not enough players'**
  String get notEnoughPlayers;

  /// No description provided for @insufficientBalanceDisqualified.
  ///
  /// In en, this message translates to:
  /// **'Insufficient balance, you cannot participate in this round'**
  String get insufficientBalanceDisqualified;

  /// No description provided for @disqualifiedFromRound.
  ///
  /// In en, this message translates to:
  /// **'You have been disqualified from this round'**
  String get disqualifiedFromRound;

  /// No description provided for @roundCancelledNotEnoughPlayers.
  ///
  /// In en, this message translates to:
  /// **'Round cancelled: need {minRequired} players, only {currentPlayers} qualified'**
  String roundCancelledNotEnoughPlayers(int minRequired, int currentPlayers);

  /// No description provided for @gameRecords.
  ///
  /// In en, this message translates to:
  /// **'Records'**
  String get gameRecords;

  /// No description provided for @myProfile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get myProfile;

  /// No description provided for @owner.
  ///
  /// In en, this message translates to:
  /// **'Owner'**
  String get owner;

  /// No description provided for @admin.
  ///
  /// In en, this message translates to:
  /// **'Admin'**
  String get admin;

  /// No description provided for @player.
  ///
  /// In en, this message translates to:
  /// **'Player'**
  String get player;

  /// No description provided for @commission.
  ///
  /// In en, this message translates to:
  /// **'Commission'**
  String get commission;

  /// No description provided for @noRooms.
  ///
  /// In en, this message translates to:
  /// **'No rooms available'**
  String get noRooms;

  /// No description provided for @contactOwnerToCreate.
  ///
  /// In en, this message translates to:
  /// **'Contact owner to create a room'**
  String get contactOwnerToCreate;

  /// No description provided for @enterRoomPassword.
  ///
  /// In en, this message translates to:
  /// **'Enter room password'**
  String get enterRoomPassword;

  /// No description provided for @passwordHint.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get passwordHint;

  /// No description provided for @availableBalance.
  ///
  /// In en, this message translates to:
  /// **'Available'**
  String get availableBalance;

  /// No description provided for @totalBalance.
  ///
  /// In en, this message translates to:
  /// **'Total'**
  String get totalBalance;

  /// No description provided for @marginDeposit.
  ///
  /// In en, this message translates to:
  /// **'Margin'**
  String get marginDeposit;

  /// No description provided for @commissionEarnings.
  ///
  /// In en, this message translates to:
  /// **'Commission'**
  String get commissionEarnings;

  /// No description provided for @myInviteCode.
  ///
  /// In en, this message translates to:
  /// **'My Invite Code'**
  String get myInviteCode;

  /// No description provided for @inviteCodeCopied.
  ///
  /// In en, this message translates to:
  /// **'Invite code copied'**
  String get inviteCodeCopied;

  /// No description provided for @selectLanguage.
  ///
  /// In en, this message translates to:
  /// **'Select Language'**
  String get selectLanguage;

  /// No description provided for @confirmLogout.
  ///
  /// In en, this message translates to:
  /// **'Confirm Logout'**
  String get confirmLogout;

  /// No description provided for @confirmLogoutMessage.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to log out?'**
  String get confirmLogoutMessage;

  /// No description provided for @loadFailed.
  ///
  /// In en, this message translates to:
  /// **'Load failed'**
  String get loadFailed;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// No description provided for @noFriends.
  ///
  /// In en, this message translates to:
  /// **'No friends yet'**
  String get noFriends;

  /// No description provided for @clickToAddFriend.
  ///
  /// In en, this message translates to:
  /// **'Click top right to add friends'**
  String get clickToAddFriend;

  /// No description provided for @searchFriends.
  ///
  /// In en, this message translates to:
  /// **'Search friends...'**
  String get searchFriends;

  /// No description provided for @addFriend.
  ///
  /// In en, this message translates to:
  /// **'Add Friend'**
  String get addFriend;

  /// No description provided for @userId.
  ///
  /// In en, this message translates to:
  /// **'User ID'**
  String get userId;

  /// No description provided for @enterUserId.
  ///
  /// In en, this message translates to:
  /// **'Enter user ID'**
  String get enterUserId;

  /// No description provided for @send.
  ///
  /// In en, this message translates to:
  /// **'Send'**
  String get send;

  /// No description provided for @friendRequestSent.
  ///
  /// In en, this message translates to:
  /// **'Friend request sent'**
  String get friendRequestSent;

  /// No description provided for @sendFailed.
  ///
  /// In en, this message translates to:
  /// **'Send failed'**
  String get sendFailed;

  /// No description provided for @inRoom.
  ///
  /// In en, this message translates to:
  /// **'In room'**
  String get inRoom;

  /// No description provided for @inviteToRoom.
  ///
  /// In en, this message translates to:
  /// **'Invite to room'**
  String get inviteToRoom;

  /// No description provided for @removeFriend.
  ///
  /// In en, this message translates to:
  /// **'Remove friend'**
  String get removeFriend;

  /// No description provided for @confirmRemoveFriend.
  ///
  /// In en, this message translates to:
  /// **'Remove Friend'**
  String get confirmRemoveFriend;

  /// No description provided for @confirmRemoveFriendMessage.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to remove {name}?'**
  String confirmRemoveFriendMessage(String name);

  /// No description provided for @remove.
  ///
  /// In en, this message translates to:
  /// **'Remove'**
  String get remove;

  /// No description provided for @totalGames.
  ///
  /// In en, this message translates to:
  /// **'Games'**
  String get totalGames;

  /// No description provided for @wins.
  ///
  /// In en, this message translates to:
  /// **'Wins'**
  String get wins;

  /// No description provided for @winRate.
  ///
  /// In en, this message translates to:
  /// **'Win Rate'**
  String get winRate;

  /// No description provided for @totalBet.
  ///
  /// In en, this message translates to:
  /// **'Total Bet'**
  String get totalBet;

  /// No description provided for @totalWon.
  ///
  /// In en, this message translates to:
  /// **'Total Won'**
  String get totalWon;

  /// No description provided for @netProfit.
  ///
  /// In en, this message translates to:
  /// **'Net P/L'**
  String get netProfit;

  /// No description provided for @roundNumber.
  ///
  /// In en, this message translates to:
  /// **'Round {number}'**
  String roundNumber(int number);

  /// No description provided for @skipped.
  ///
  /// In en, this message translates to:
  /// **'Skipped'**
  String get skipped;

  /// No description provided for @bet.
  ///
  /// In en, this message translates to:
  /// **'Bet'**
  String get bet;

  /// No description provided for @ownerWallet.
  ///
  /// In en, this message translates to:
  /// **'Owner Wallet'**
  String get ownerWallet;

  /// No description provided for @marginBalance.
  ///
  /// In en, this message translates to:
  /// **'Margin'**
  String get marginBalance;

  /// No description provided for @commissionBalance.
  ///
  /// In en, this message translates to:
  /// **'Commission'**
  String get commissionBalance;

  /// No description provided for @totalCommission.
  ///
  /// In en, this message translates to:
  /// **'Total Commission'**
  String get totalCommission;

  /// No description provided for @canRechargePlayer.
  ///
  /// In en, this message translates to:
  /// **'For player recharge'**
  String get canRechargePlayer;

  /// No description provided for @fixedGuarantee.
  ///
  /// In en, this message translates to:
  /// **'Fixed guarantee'**
  String get fixedGuarantee;

  /// No description provided for @canTransferToBalance.
  ///
  /// In en, this message translates to:
  /// **'Can transfer to balance'**
  String get canTransferToBalance;

  /// No description provided for @transferToBalance.
  ///
  /// In en, this message translates to:
  /// **'Transfer to Balance'**
  String get transferToBalance;

  /// No description provided for @transferAmount.
  ///
  /// In en, this message translates to:
  /// **'Transfer Amount'**
  String get transferAmount;

  /// No description provided for @maxTransfer.
  ///
  /// In en, this message translates to:
  /// **'Max {amount}'**
  String maxTransfer(String amount);

  /// No description provided for @transfer.
  ///
  /// In en, this message translates to:
  /// **'Transfer'**
  String get transfer;

  /// No description provided for @transferNote.
  ///
  /// In en, this message translates to:
  /// **'Note: Transfer commission to available balance for player recharge or withdrawal'**
  String get transferNote;

  /// No description provided for @submitFundRequest.
  ///
  /// In en, this message translates to:
  /// **'Submit Fund Request'**
  String get submitFundRequest;

  /// No description provided for @ownerDeposit.
  ///
  /// In en, this message translates to:
  /// **'Balance Deposit'**
  String get ownerDeposit;

  /// No description provided for @ownerWithdraw.
  ///
  /// In en, this message translates to:
  /// **'Balance Withdraw'**
  String get ownerWithdraw;

  /// No description provided for @marginDeposit2.
  ///
  /// In en, this message translates to:
  /// **'Margin Deposit'**
  String get marginDeposit2;

  /// No description provided for @amount.
  ///
  /// In en, this message translates to:
  /// **'Amount'**
  String get amount;

  /// No description provided for @remarkOptional.
  ///
  /// In en, this message translates to:
  /// **'Remark (Optional)'**
  String get remarkOptional;

  /// No description provided for @submitRequest.
  ///
  /// In en, this message translates to:
  /// **'Submit Request'**
  String get submitRequest;

  /// No description provided for @fundRequestNote.
  ///
  /// In en, this message translates to:
  /// **'Note: Deposit/withdrawal requests need admin approval before balance changes'**
  String get fundRequestNote;

  /// No description provided for @fundRequestRecords.
  ///
  /// In en, this message translates to:
  /// **'Fund Request Records'**
  String get fundRequestRecords;

  /// No description provided for @pending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get pending;

  /// No description provided for @approved.
  ///
  /// In en, this message translates to:
  /// **'Approved'**
  String get approved;

  /// No description provided for @rejected.
  ///
  /// In en, this message translates to:
  /// **'Rejected'**
  String get rejected;

  /// No description provided for @transactionHistory.
  ///
  /// In en, this message translates to:
  /// **'Transaction History'**
  String get transactionHistory;

  /// No description provided for @gameBet.
  ///
  /// In en, this message translates to:
  /// **'Game Bet'**
  String get gameBet;

  /// No description provided for @gameWin.
  ///
  /// In en, this message translates to:
  /// **'Game Win'**
  String get gameWin;

  /// No description provided for @gameRefund.
  ///
  /// In en, this message translates to:
  /// **'Game Refund'**
  String get gameRefund;

  /// No description provided for @ownerCommission.
  ///
  /// In en, this message translates to:
  /// **'Owner Commission'**
  String get ownerCommission;

  /// No description provided for @platformShare.
  ///
  /// In en, this message translates to:
  /// **'Platform Share'**
  String get platformShare;

  /// No description provided for @pleaseEnterAmount.
  ///
  /// In en, this message translates to:
  /// **'Please enter amount'**
  String get pleaseEnterAmount;

  /// No description provided for @invalidAmountFormat.
  ///
  /// In en, this message translates to:
  /// **'Invalid amount format'**
  String get invalidAmountFormat;

  /// No description provided for @pleaseSelectType.
  ///
  /// In en, this message translates to:
  /// **'Please select type'**
  String get pleaseSelectType;

  /// No description provided for @fundRequestSubmitted.
  ///
  /// In en, this message translates to:
  /// **'Fund request submitted, pending approval'**
  String get fundRequestSubmitted;

  /// No description provided for @submitFailed.
  ///
  /// In en, this message translates to:
  /// **'Submit failed'**
  String get submitFailed;

  /// No description provided for @pleaseEnterTransferAmount.
  ///
  /// In en, this message translates to:
  /// **'Please enter transfer amount'**
  String get pleaseEnterTransferAmount;

  /// No description provided for @transferAmountExceeds.
  ///
  /// In en, this message translates to:
  /// **'Transfer amount cannot exceed ¥{amount}'**
  String transferAmountExceeds(String amount);

  /// No description provided for @transferSuccess.
  ///
  /// In en, this message translates to:
  /// **'Transfer successful'**
  String get transferSuccess;

  /// No description provided for @transferFailed.
  ///
  /// In en, this message translates to:
  /// **'Transfer failed'**
  String get transferFailed;

  /// No description provided for @pendingPlayerRequests.
  ///
  /// In en, this message translates to:
  /// **'Pending Player Requests'**
  String get pendingPlayerRequests;

  /// No description provided for @noPendingRequests.
  ///
  /// In en, this message translates to:
  /// **'No pending requests'**
  String get noPendingRequests;

  /// No description provided for @unknownUser.
  ///
  /// In en, this message translates to:
  /// **'Unknown User'**
  String get unknownUser;

  /// No description provided for @confirmApprove.
  ///
  /// In en, this message translates to:
  /// **'Confirm Approve'**
  String get confirmApprove;

  /// No description provided for @confirmReject.
  ///
  /// In en, this message translates to:
  /// **'Confirm Reject'**
  String get confirmReject;

  /// No description provided for @confirmApproveMessage.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to approve this request?'**
  String get confirmApproveMessage;

  /// No description provided for @confirmRejectMessage.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to reject this request?'**
  String get confirmRejectMessage;

  /// No description provided for @approve.
  ///
  /// In en, this message translates to:
  /// **'Approve'**
  String get approve;

  /// No description provided for @reject.
  ///
  /// In en, this message translates to:
  /// **'Reject'**
  String get reject;

  /// No description provided for @requestApproved.
  ///
  /// In en, this message translates to:
  /// **'Request approved'**
  String get requestApproved;

  /// No description provided for @requestRejected.
  ///
  /// In en, this message translates to:
  /// **'Request rejected'**
  String get requestRejected;

  /// No description provided for @operationFailed.
  ///
  /// In en, this message translates to:
  /// **'Operation failed'**
  String get operationFailed;

  /// No description provided for @basicInfo.
  ///
  /// In en, this message translates to:
  /// **'Basic Info'**
  String get basicInfo;

  /// No description provided for @gameSettings.
  ///
  /// In en, this message translates to:
  /// **'Game Settings'**
  String get gameSettings;

  /// No description provided for @feeSettings.
  ///
  /// In en, this message translates to:
  /// **'Fee Settings'**
  String get feeSettings;

  /// No description provided for @securitySettings.
  ///
  /// In en, this message translates to:
  /// **'Security Settings'**
  String get securitySettings;

  /// No description provided for @platformFeeFixed.
  ///
  /// In en, this message translates to:
  /// **'Platform Fee (Fixed)'**
  String get platformFeeFixed;

  /// No description provided for @ownerFee.
  ///
  /// In en, this message translates to:
  /// **'Owner Fee'**
  String get ownerFee;

  /// No description provided for @totalFee.
  ///
  /// In en, this message translates to:
  /// **'Total Fee'**
  String get totalFee;

  /// No description provided for @adjustOwnerFee.
  ///
  /// In en, this message translates to:
  /// **'Adjust Owner Fee:'**
  String get adjustOwnerFee;

  /// No description provided for @feeNote.
  ///
  /// In en, this message translates to:
  /// **'Note: Total fee is deducted from prize pool at settlement'**
  String get feeNote;

  /// No description provided for @roomPassword.
  ///
  /// In en, this message translates to:
  /// **'Room Password'**
  String get roomPassword;

  /// No description provided for @setPassword.
  ///
  /// In en, this message translates to:
  /// **'Set Password'**
  String get setPassword;

  /// No description provided for @pleaseEnterRoomName.
  ///
  /// In en, this message translates to:
  /// **'Please enter room name'**
  String get pleaseEnterRoomName;

  /// No description provided for @pleaseEnterPassword.
  ///
  /// In en, this message translates to:
  /// **'Please enter password'**
  String get pleaseEnterPassword;

  /// No description provided for @createFailed.
  ///
  /// In en, this message translates to:
  /// **'Create failed'**
  String get createFailed;

  /// No description provided for @winnersMustLessThanPlayers.
  ///
  /// In en, this message translates to:
  /// **'Winners must be less than max players'**
  String get winnersMustLessThanPlayers;

  /// No description provided for @minPlayersHint.
  ///
  /// In en, this message translates to:
  /// **'Winners per round, need at least {count} players to start'**
  String minPlayersHint(int count);

  /// No description provided for @maxPlayersHint.
  ///
  /// In en, this message translates to:
  /// **'Maximum players in room'**
  String get maxPlayersHint;

  /// No description provided for @connectionError.
  ///
  /// In en, this message translates to:
  /// **'Connection error'**
  String get connectionError;

  /// No description provided for @connectionClosed.
  ///
  /// In en, this message translates to:
  /// **'Connection closed, reconnecting...'**
  String get connectionClosed;

  /// No description provided for @joinRoomFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to join room'**
  String get joinRoomFailed;

  /// No description provided for @joinRoomTitle.
  ///
  /// In en, this message translates to:
  /// **'Join Room'**
  String get joinRoomTitle;

  /// No description provided for @currentBalance.
  ///
  /// In en, this message translates to:
  /// **'Current Balance'**
  String get currentBalance;

  /// No description provided for @minBalanceRequired.
  ///
  /// In en, this message translates to:
  /// **'Min Balance Required'**
  String get minBalanceRequired;

  /// No description provided for @insufficientBalanceCannotJoin.
  ///
  /// In en, this message translates to:
  /// **'Insufficient balance, cannot join room'**
  String get insufficientBalanceCannotJoin;

  /// No description provided for @riskWarning.
  ///
  /// In en, this message translates to:
  /// **'Risk Warning'**
  String get riskWarning;

  /// No description provided for @riskWarningTip1.
  ///
  /// In en, this message translates to:
  /// **'Gambling involves risk, please participate responsibly'**
  String get riskWarningTip1;

  /// No description provided for @riskWarningTip2.
  ///
  /// In en, this message translates to:
  /// **'You will automatically join each round after entering'**
  String get riskWarningTip2;

  /// No description provided for @confirmJoin.
  ///
  /// In en, this message translates to:
  /// **'Confirm Join'**
  String get confirmJoin;

  /// No description provided for @minPlayersRequired.
  ///
  /// In en, this message translates to:
  /// **'Min Players'**
  String get minPlayersRequired;

  /// No description provided for @winnersPerRound.
  ///
  /// In en, this message translates to:
  /// **'Winners/Round'**
  String get winnersPerRound;

  /// No description provided for @maxPlayersAllowed.
  ///
  /// In en, this message translates to:
  /// **'Max Players'**
  String get maxPlayersAllowed;

  /// No description provided for @betPerRoundLabel.
  ///
  /// In en, this message translates to:
  /// **'Bet/Round'**
  String get betPerRoundLabel;

  /// No description provided for @personUnit.
  ///
  /// In en, this message translates to:
  /// **''**
  String get personUnit;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'ja', 'ko', 'zh'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when language+country codes are specified.
  switch (locale.languageCode) {
    case 'zh':
      {
        switch (locale.countryCode) {
          case 'TW':
            return AppLocalizationsZhTw();
        }
        break;
      }
  }

  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'ja':
      return AppLocalizationsJa();
    case 'ko':
      return AppLocalizationsKo();
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
