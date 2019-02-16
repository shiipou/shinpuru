CREATE TABLE IF NOT EXISTS `guilds` (
  `guildID` text NOT NULL DEFAULT '',
  `prefix` text NOT NULL DEFAULT '',
  `autorole` text NOT NULL DEFAULT '',
  `modlogchanID` text NOT NULL DEFAULT '',
  `voicelogchanID` text NOT NULL DEFAULT '',
  `muteRoleID` text NOT NULL DEFAULT '',
  `ghostPingMsg` text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS `permissions` (
  `roleID` text NOT NULL DEFAULT '',
  `guildID` text NOT NULL DEFAULT '',
  `permission` int(11) NOT NULL DEFAULT '0'
);

CREATE TABLE IF NOT EXISTS `reports` (
  `id` text NOT NULL DEFAULT '',
  `type` int(11) NOT NULL DEFAULT '3',
  `guildID` text NOT NULL DEFAULT '',
  `executorID` text NOT NULL DEFAULT '',
  `victimID` text NOT NULL DEFAULT '',
  `msg` text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS `settings` (
  `setting` text NOT NULL DEFAULT '',
  `value` text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS `starboard` (
  `guildID` text NOT NULL DEFAULT '',
  `chanID` text NOT NULL DEFAULT '',
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `minimum` int(11) NOT NULL DEFAULT '5'
);

CREATE TABLE IF NOT EXISTS `votes` (
  `ID` text NOT NULL DEFAULT '',
  `data` mediumtext NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS `twitchnotify` (
  `guildID` text NOT NULL DEFAULT '',
  `channelID` text NOT NULL DEFAULT '',
  `twitchUserID` text NOT NULL DEFAULT ''
);