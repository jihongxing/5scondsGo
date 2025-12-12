import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
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
    Locale('zh')
  ];

  /// No description provided for @appTitle.
  ///
  /// In en, this message translates to:
  /// **'5SecondsGo Admin'**
  String get appTitle;

  /// No description provided for @navDashboard.
  ///
  /// In en, this message translates to:
  /// **'Dashboard'**
  String get navDashboard;

  /// No description provided for @navUsers.
  ///
  /// In en, this message translates to:
  /// **'Users'**
  String get navUsers;

  /// No description provided for @navRooms.
  ///
  /// In en, this message translates to:
  /// **'Rooms'**
  String get navRooms;

  /// No description provided for @navFunds.
  ///
  /// In en, this message translates to:
  /// **'Funds'**
  String get navFunds;

  /// No description provided for @navLogout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get navLogout;

  /// No description provided for @loginTitle.
  ///
  /// In en, this message translates to:
  /// **'5SecondsGo Admin'**
  String get loginTitle;

  /// No description provided for @loginUsernameLabel.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get loginUsernameLabel;

  /// No description provided for @loginUsernameRequired.
  ///
  /// In en, this message translates to:
  /// **'Please enter username'**
  String get loginUsernameRequired;

  /// No description provided for @loginPasswordLabel.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get loginPasswordLabel;

  /// No description provided for @loginPasswordTooShort.
  ///
  /// In en, this message translates to:
  /// **'Password must be at least 6 characters'**
  String get loginPasswordTooShort;

  /// No description provided for @loginButton.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get loginButton;

  /// No description provided for @dashboardTitle.
  ///
  /// In en, this message translates to:
  /// **'Dashboard'**
  String get dashboardTitle;

  /// No description provided for @dashboardCardTotalUsers.
  ///
  /// In en, this message translates to:
  /// **'Total Users'**
  String get dashboardCardTotalUsers;

  /// No description provided for @dashboardCardActiveRooms.
  ///
  /// In en, this message translates to:
  /// **'Active Rooms'**
  String get dashboardCardActiveRooms;

  /// No description provided for @dashboardCardOnlinePlayers.
  ///
  /// In en, this message translates to:
  /// **'Online Players'**
  String get dashboardCardOnlinePlayers;

  /// No description provided for @dashboardCardPlatformBalance.
  ///
  /// In en, this message translates to:
  /// **'Platform Balance'**
  String get dashboardCardPlatformBalance;

  /// No description provided for @dashboardFundCheckTitle.
  ///
  /// In en, this message translates to:
  /// **'Fund Conservation Check'**
  String get dashboardFundCheckTitle;

  /// No description provided for @dashboardFundCheckPlayerBalance.
  ///
  /// In en, this message translates to:
  /// **'Total player balance: {amount}'**
  String dashboardFundCheckPlayerBalance(Object amount);

  /// No description provided for @dashboardFundCheckCustodyQuota.
  ///
  /// In en, this message translates to:
  /// **'Total owner balance: {amount}'**
  String dashboardFundCheckCustodyQuota(Object amount);

  /// No description provided for @dashboardFundCheckMargin.
  ///
  /// In en, this message translates to:
  /// **'Total owner profit: {amount}'**
  String dashboardFundCheckMargin(Object amount);

  /// No description provided for @dashboardFundCheckPlatformProfit.
  ///
  /// In en, this message translates to:
  /// **'Total platform profit: {amount}'**
  String dashboardFundCheckPlatformProfit(Object amount);

  /// No description provided for @dashboardFundCheckBalanced.
  ///
  /// In en, this message translates to:
  /// **'✓ Balanced'**
  String get dashboardFundCheckBalanced;

  /// No description provided for @dashboardRecentFundRequests.
  ///
  /// In en, this message translates to:
  /// **'Recent Fund Requests'**
  String get dashboardRecentFundRequests;

  /// No description provided for @dashboardRecentFundTitle.
  ///
  /// In en, this message translates to:
  /// **'{username} - {type}'**
  String dashboardRecentFundTitle(Object type, Object username);

  /// No description provided for @dashboardRecentFundTypeDeposit.
  ///
  /// In en, this message translates to:
  /// **'Deposit'**
  String get dashboardRecentFundTypeDeposit;

  /// No description provided for @dashboardRecentFundTypeWithdraw.
  ///
  /// In en, this message translates to:
  /// **'Withdraw'**
  String get dashboardRecentFundTypeWithdraw;

  /// No description provided for @dashboardRecentFundStatusPending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get dashboardRecentFundStatusPending;

  /// No description provided for @dashboardRecentFundStatusApproved.
  ///
  /// In en, this message translates to:
  /// **'Approved'**
  String get dashboardRecentFundStatusApproved;

  /// No description provided for @dashboardRecentFundStatusRejected.
  ///
  /// In en, this message translates to:
  /// **'Rejected'**
  String get dashboardRecentFundStatusRejected;

  /// No description provided for @dashboardAdminInviteTitle.
  ///
  /// In en, this message translates to:
  /// **'Admin Account'**
  String get dashboardAdminInviteTitle;

  /// No description provided for @dashboardAdminInviteCodeLabel.
  ///
  /// In en, this message translates to:
  /// **'Invite code: {code}'**
  String dashboardAdminInviteCodeLabel(Object code);

  /// No description provided for @fundsTitle.
  ///
  /// In en, this message translates to:
  /// **'Funds'**
  String get fundsTitle;

  /// No description provided for @fundsTabRequests.
  ///
  /// In en, this message translates to:
  /// **'Fund Requests'**
  String get fundsTabRequests;

  /// No description provided for @fundsTabTransactions.
  ///
  /// In en, this message translates to:
  /// **'Transactions'**
  String get fundsTabTransactions;

  /// No description provided for @fundsFilterAllStatus.
  ///
  /// In en, this message translates to:
  /// **'All Status'**
  String get fundsFilterAllStatus;

  /// No description provided for @fundsFilterStatusPending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get fundsFilterStatusPending;

  /// No description provided for @fundsFilterStatusApproved.
  ///
  /// In en, this message translates to:
  /// **'Approved'**
  String get fundsFilterStatusApproved;

  /// No description provided for @fundsFilterStatusRejected.
  ///
  /// In en, this message translates to:
  /// **'Rejected'**
  String get fundsFilterStatusRejected;

  /// No description provided for @fundsFilterAllTypes.
  ///
  /// In en, this message translates to:
  /// **'All Types'**
  String get fundsFilterAllTypes;

  /// No description provided for @fundsFilterTypeDeposit.
  ///
  /// In en, this message translates to:
  /// **'Deposit'**
  String get fundsFilterTypeDeposit;

  /// No description provided for @fundsFilterTypeWithdraw.
  ///
  /// In en, this message translates to:
  /// **'Withdraw'**
  String get fundsFilterTypeWithdraw;

  /// No description provided for @fundsFilterTypeMargin.
  ///
  /// In en, this message translates to:
  /// **'Margin'**
  String get fundsFilterTypeMargin;

  /// No description provided for @fundsStatusPending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get fundsStatusPending;

  /// No description provided for @fundsStatusApproved.
  ///
  /// In en, this message translates to:
  /// **'Approved'**
  String get fundsStatusApproved;

  /// No description provided for @fundsStatusRejected.
  ///
  /// In en, this message translates to:
  /// **'Rejected'**
  String get fundsStatusRejected;

  /// No description provided for @fundsSearchUserHint.
  ///
  /// In en, this message translates to:
  /// **'Search by user...'**
  String get fundsSearchUserHint;

  /// No description provided for @fundsTxTypeGameBet.
  ///
  /// In en, this message translates to:
  /// **'Game Bet'**
  String get fundsTxTypeGameBet;

  /// No description provided for @fundsTxTypeGameWin.
  ///
  /// In en, this message translates to:
  /// **'Game Win'**
  String get fundsTxTypeGameWin;

  /// No description provided for @fundsTxTypeDeposit.
  ///
  /// In en, this message translates to:
  /// **'Deposit'**
  String get fundsTxTypeDeposit;

  /// No description provided for @fundsTxTypeWithdraw.
  ///
  /// In en, this message translates to:
  /// **'Withdraw'**
  String get fundsTxTypeWithdraw;

  /// No description provided for @roomsTitle.
  ///
  /// In en, this message translates to:
  /// **'Rooms'**
  String get roomsTitle;

  /// No description provided for @roomsSearchHint.
  ///
  /// In en, this message translates to:
  /// **'Search rooms...'**
  String get roomsSearchHint;

  /// No description provided for @roomsFilterAllStatus.
  ///
  /// In en, this message translates to:
  /// **'All Status'**
  String get roomsFilterAllStatus;

  /// No description provided for @roomsFilterStatusActive.
  ///
  /// In en, this message translates to:
  /// **'Active'**
  String get roomsFilterStatusActive;

  /// No description provided for @roomsFilterStatusPaused.
  ///
  /// In en, this message translates to:
  /// **'Paused'**
  String get roomsFilterStatusPaused;

  /// No description provided for @roomsFilterStatusLocked.
  ///
  /// In en, this message translates to:
  /// **'Locked'**
  String get roomsFilterStatusLocked;

  /// No description provided for @roomsItemTitle.
  ///
  /// In en, this message translates to:
  /// **'{roomName}'**
  String roomsItemTitle(Object roomName);

  /// No description provided for @roomsItemSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Owner: {ownerName} · Players: {players}/20 · Bet: {bet}'**
  String roomsItemSubtitle(Object bet, Object ownerName, Object players);

  /// No description provided for @roomsStatusActive.
  ///
  /// In en, this message translates to:
  /// **'ACTIVE'**
  String get roomsStatusActive;

  /// No description provided for @roomsStatusPaused.
  ///
  /// In en, this message translates to:
  /// **'PAUSED'**
  String get roomsStatusPaused;

  /// No description provided for @roomsStatusLocked.
  ///
  /// In en, this message translates to:
  /// **'LOCKED'**
  String get roomsStatusLocked;

  /// No description provided for @usersTitle.
  ///
  /// In en, this message translates to:
  /// **'Users'**
  String get usersTitle;

  /// No description provided for @usersCreateOwner.
  ///
  /// In en, this message translates to:
  /// **'Create Owner'**
  String get usersCreateOwner;

  /// No description provided for @usersSearchHint.
  ///
  /// In en, this message translates to:
  /// **'Search users...'**
  String get usersSearchHint;

  /// No description provided for @usersFilterAllRoles.
  ///
  /// In en, this message translates to:
  /// **'All Roles'**
  String get usersFilterAllRoles;

  /// No description provided for @usersFilterRolePlayer.
  ///
  /// In en, this message translates to:
  /// **'Player'**
  String get usersFilterRolePlayer;

  /// No description provided for @usersFilterRoleOwner.
  ///
  /// In en, this message translates to:
  /// **'Owner'**
  String get usersFilterRoleOwner;

  /// No description provided for @usersFilterRoleAdmin.
  ///
  /// In en, this message translates to:
  /// **'Admin'**
  String get usersFilterRoleAdmin;

  /// No description provided for @usersItemTitle.
  ///
  /// In en, this message translates to:
  /// **'{username}'**
  String usersItemTitle(Object username);

  /// No description provided for @usersItemSubtitleOwner.
  ///
  /// In en, this message translates to:
  /// **'Owner · Invite: {code}'**
  String usersItemSubtitleOwner(Object code);

  /// No description provided for @usersItemSubtitlePlayer.
  ///
  /// In en, this message translates to:
  /// **'Player · Invited by: {ownerName}'**
  String usersItemSubtitlePlayer(Object ownerName);

  /// No description provided for @usersDialogCreateOwnerTitle.
  ///
  /// In en, this message translates to:
  /// **'Create Owner'**
  String get usersDialogCreateOwnerTitle;

  /// No description provided for @usersDialogUsernameLabel.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get usersDialogUsernameLabel;

  /// No description provided for @usersDialogPasswordLabel.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get usersDialogPasswordLabel;

  /// No description provided for @usersDialogCancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get usersDialogCancel;

  /// No description provided for @usersDialogCreate.
  ///
  /// In en, this message translates to:
  /// **'Create'**
  String get usersDialogCreate;

  /// No description provided for @navMonitoring.
  ///
  /// In en, this message translates to:
  /// **'Monitoring'**
  String get navMonitoring;

  /// No description provided for @navRiskFlags.
  ///
  /// In en, this message translates to:
  /// **'Risk Flags'**
  String get navRiskFlags;

  /// No description provided for @navAlerts.
  ///
  /// In en, this message translates to:
  /// **'Alerts'**
  String get navAlerts;

  /// No description provided for @monitoringTitle.
  ///
  /// In en, this message translates to:
  /// **'Monitoring Dashboard'**
  String get monitoringTitle;

  /// No description provided for @monitoringRealtimeMetrics.
  ///
  /// In en, this message translates to:
  /// **'Realtime Metrics'**
  String get monitoringRealtimeMetrics;

  /// No description provided for @monitoringPerformanceMetrics.
  ///
  /// In en, this message translates to:
  /// **'Performance Metrics'**
  String get monitoringPerformanceMetrics;

  /// No description provided for @monitoringHistoricalTrends.
  ///
  /// In en, this message translates to:
  /// **'Historical Trends'**
  String get monitoringHistoricalTrends;

  /// No description provided for @monitoringOnlinePlayers.
  ///
  /// In en, this message translates to:
  /// **'Online Players'**
  String get monitoringOnlinePlayers;

  /// No description provided for @monitoringActiveRooms.
  ///
  /// In en, this message translates to:
  /// **'Active Rooms'**
  String get monitoringActiveRooms;

  /// No description provided for @monitoringGamesPerMinute.
  ///
  /// In en, this message translates to:
  /// **'Games/Min'**
  String get monitoringGamesPerMinute;

  /// No description provided for @monitoringDailyActiveUsers.
  ///
  /// In en, this message translates to:
  /// **'Daily Active Users'**
  String get monitoringDailyActiveUsers;

  /// No description provided for @monitoringDailyVolume.
  ///
  /// In en, this message translates to:
  /// **'Daily Volume'**
  String get monitoringDailyVolume;

  /// No description provided for @monitoringPlatformRevenue.
  ///
  /// In en, this message translates to:
  /// **'Platform Revenue'**
  String get monitoringPlatformRevenue;

  /// No description provided for @monitoringApiLatencyP95.
  ///
  /// In en, this message translates to:
  /// **'API Latency P95'**
  String get monitoringApiLatencyP95;

  /// No description provided for @monitoringWsLatencyP95.
  ///
  /// In en, this message translates to:
  /// **'WS Latency P95'**
  String get monitoringWsLatencyP95;

  /// No description provided for @monitoringDbLatencyP95.
  ///
  /// In en, this message translates to:
  /// **'DB Latency P95'**
  String get monitoringDbLatencyP95;

  /// No description provided for @refresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh'**
  String get refresh;
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
      <String>['en', 'zh'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
