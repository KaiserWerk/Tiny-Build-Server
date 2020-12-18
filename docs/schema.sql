-- phpMyAdmin SQL Dump
-- version 4.9.5
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Erstellungszeit: 18. Dez 2020 um 21:46
-- Server-Version: 5.7.24
-- PHP-Version: 7.2.14

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Datenbank: `tiny_build_server`
--

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_definition`
--

CREATE TABLE `build_definition` (
  `id` int(10) UNSIGNED NOT NULL,
  `build_target` varchar(25) NOT NULL,
  `build_target_os_arch` varchar(100) NOT NULL DEFAULT '',
  `build_target_arm` int(11) NOT NULL DEFAULT '0',
  `altered_by` int(10) UNSIGNED NOT NULL,
  `caption` varchar(75) NOT NULL DEFAULT '',
  `enabled` tinyint(1) UNSIGNED NOT NULL DEFAULT '1',
  `deployment_enabled` tinyint(1) UNSIGNED NOT NULL DEFAULT '1',
  `repo_hoster` varchar(15) NOT NULL,
  `repo_hoster_url` varchar(200) NOT NULL,
  `repo_fullname` varchar(150) NOT NULL,
  `repo_username` varchar(100) NOT NULL DEFAULT '',
  `repo_secret` varchar(150) NOT NULL DEFAULT '',
  `repo_branch` varchar(150) NOT NULL,
  `altered_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `apply_migrations` tinyint(1) NOT NULL DEFAULT '0',
  `database_dsn` varchar(255) NOT NULL DEFAULT '',
  `meta_migration_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
  `run_tests` tinyint(1) NOT NULL DEFAULT '0',
  `run_benchmark_tests` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_execution`
--

CREATE TABLE `build_execution` (
  `id` int(10) UNSIGNED NOT NULL,
  `build_definition_id` int(11) NOT NULL,
  `initiated_by` int(11) NOT NULL DEFAULT '-1',
  `manual_run` tinyint(1) NOT NULL DEFAULT '0',
  `action_log` mediumtext NOT NULL,
  `result` varchar(40) NOT NULL,
  `artifact_path` varchar(255) NOT NULL DEFAULT '',
  `execution_time` decimal(10,2) NOT NULL,
  `executed_at` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_request`
--

CREATE TABLE `build_request` (
  `id` int(11) NOT NULL,
  `build_definition_id` int(11) DEFAULT NULL,
  `headers` text NOT NULL,
  `payload` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_step`
--

CREATE TABLE `build_step` (
  `id` int(11) NOT NULL,
  `build_target_id` int(11) NOT NULL,
  `internal_id` varchar(30) NOT NULL,
  `caption` varchar(150) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_target`
--

CREATE TABLE `build_target` (
  `id` int(11) NOT NULL,
  `caption` varchar(150) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_variable`
--

CREATE TABLE `build_variable` (
  `id` int(11) UNSIGNED NOT NULL,
  `user_id` int(11) NOT NULL,
  `description` varchar(150) NOT NULL,
  `content` text NOT NULL,
  `user_specific` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `deployment_definition`
--

CREATE TABLE `deployment_definition` (
  `id` int(10) UNSIGNED NOT NULL,
  `build_definition_id` int(10) UNSIGNED NOT NULL,
  `caption` varchar(100) NOT NULL,
  `host` varchar(200) NOT NULL,
  `username` varchar(100) NOT NULL,
  `password` varchar(150) NOT NULL,
  `connection_type` varchar(10) NOT NULL,
  `working_directory` varchar(250) NOT NULL,
  `pre_deployment_actions` text,
  `post_deployment_actions` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `setting`
--

CREATE TABLE `setting` (
  `id` int(11) NOT NULL,
  `setting_name` varchar(150) NOT NULL,
  `setting_value` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `user`
--

CREATE TABLE `user` (
  `id` int(11) UNSIGNED NOT NULL,
  `displayname` varchar(50) NOT NULL,
  `email` varchar(150) NOT NULL,
  `password` varchar(150) NOT NULL,
  `locked` tinyint(1) NOT NULL DEFAULT '0',
  `admin` tinyint(4) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `user_action`
--

CREATE TABLE `user_action` (
  `id` int(10) UNSIGNED NOT NULL,
  `user_id` int(10) UNSIGNED NOT NULL,
  `purpose` varchar(30) NOT NULL,
  `token` varchar(150) NOT NULL,
  `validity` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Indizes der exportierten Tabellen
--

--
-- Indizes für die Tabelle `build_definition`
--
ALTER TABLE `build_definition`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `build_execution`
--
ALTER TABLE `build_execution`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `build_request`
--
ALTER TABLE `build_request`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `build_step`
--
ALTER TABLE `build_step`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `build_target`
--
ALTER TABLE `build_target`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `build_variable`
--
ALTER TABLE `build_variable`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `deployment_definition`
--
ALTER TABLE `deployment_definition`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `setting`
--
ALTER TABLE `setting`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `setting_name` (`setting_name`);

--
-- Indizes für die Tabelle `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `displayname` (`displayname`),
  ADD UNIQUE KEY `email` (`email`);

--
-- Indizes für die Tabelle `user_action`
--
ALTER TABLE `user_action`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT für exportierte Tabellen
--

--
-- AUTO_INCREMENT für Tabelle `build_definition`
--
ALTER TABLE `build_definition`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_execution`
--
ALTER TABLE `build_execution`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_request`
--
ALTER TABLE `build_request`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_step`
--
ALTER TABLE `build_step`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_target`
--
ALTER TABLE `build_target`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_variable`
--
ALTER TABLE `build_variable`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `deployment_definition`
--
ALTER TABLE `deployment_definition`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `setting`
--
ALTER TABLE `setting`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `user`
--
ALTER TABLE `user`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `user_action`
--
ALTER TABLE `user_action`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
