-- phpMyAdmin SQL Dump
-- version 4.9.5
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Erstellungszeit: 18. Jan 2022 um 21:45
-- Server-Version: 5.7.24
-- PHP-Version: 7.4.1

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Datenbank: `tinybuildserver_orm`
--

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `admin_setting`
--

CREATE TABLE `admin_setting` (
                                 `name` longtext,
                                 `value` longtext
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_definition`
--

CREATE TABLE `build_definition` (
                                    `id` bigint(20) NOT NULL,
                                    `caption` longtext,
                                    `token` longtext,
                                    `content` longtext,
                                    `edited_by` bigint(20) DEFAULT NULL,
                                    `edited_at` datetime(3) DEFAULT NULL,
                                    `created_by` bigint(20) DEFAULT NULL,
                                    `created_at` datetime(3) DEFAULT NULL,
                                    `deleted` tinyint(1) NOT NULL,
                                    `updated_at` datetime(3) DEFAULT NULL,
                                    `deleted_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `build_execution`
--

CREATE TABLE `build_execution` (
                                   `id` bigint(20) NOT NULL,
                                   `build_definition_id` bigint(20) DEFAULT NULL,
                                   `manually_run_by` bigint(20) DEFAULT NULL,
                                   `action_log` longtext,
                                   `result` longtext,
                                   `artifact_path` longtext,
                                   `execution_time` double DEFAULT NULL,
                                   `executed_at` datetime(3) DEFAULT NULL,
                                   `updated_at` datetime(3) DEFAULT NULL,
                                   `created_at` datetime(3) DEFAULT NULL,
                                   `deleted_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `user`
--

CREATE TABLE `user` (
                        `id` bigint(20) NOT NULL,
                        `displayname` longtext,
                        `email` longtext,
                        `password` longtext,
                        `locked` tinyint(1) DEFAULT NULL,
                        `admin` tinyint(1) DEFAULT NULL,
                        `created_at` datetime(3) DEFAULT NULL,
                        `updated_at` datetime(3) DEFAULT NULL,
                        `display_name` longtext,
                        `deleted_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `user_action`
--

CREATE TABLE `user_action` (
                               `id` bigint(20) NOT NULL,
                               `user_id` bigint(20) DEFAULT NULL,
                               `purpose` longtext,
                               `token` longtext,
                               `validity` datetime(3) DEFAULT NULL,
                               `created_at` datetime(3) DEFAULT NULL,
                               `updated_at` datetime(3) DEFAULT NULL,
                               `deleted_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `user_variable`
--

CREATE TABLE `user_variable` (
                                 `id` bigint(20) NOT NULL,
                                 `user_entry_id` bigint(20) DEFAULT NULL,
                                 `variable` longtext,
                                 `value` longtext,
                                 `public` tinyint(1) DEFAULT NULL,
                                 `updated_at` datetime(3) DEFAULT NULL,
                                 `deleted_at` datetime(3) DEFAULT NULL,
                                 `created_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Indizes der exportierten Tabellen
--

--
-- Indizes für die Tabelle `build_definition`
--
ALTER TABLE `build_definition`
    ADD PRIMARY KEY (`id`),
    ADD KEY `idx_build_definition_deleted_at` (`deleted_at`);

--
-- Indizes für die Tabelle `build_execution`
--
ALTER TABLE `build_execution`
    ADD PRIMARY KEY (`id`),
    ADD KEY `fk_build_definition_build_executions` (`build_definition_id`),
    ADD KEY `idx_build_execution_deleted_at` (`deleted_at`);

--
-- Indizes für die Tabelle `user`
--
ALTER TABLE `user`
    ADD PRIMARY KEY (`id`),
    ADD KEY `idx_user_deleted_at` (`deleted_at`);

--
-- Indizes für die Tabelle `user_action`
--
ALTER TABLE `user_action`
    ADD PRIMARY KEY (`id`),
    ADD KEY `idx_user_action_deleted_at` (`deleted_at`);

--
-- Indizes für die Tabelle `user_variable`
--
ALTER TABLE `user_variable`
    ADD PRIMARY KEY (`id`),
    ADD KEY `idx_user_variable_deleted_at` (`deleted_at`);

--
-- AUTO_INCREMENT für exportierte Tabellen
--

--
-- AUTO_INCREMENT für Tabelle `build_definition`
--
ALTER TABLE `build_definition`
    MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `build_execution`
--
ALTER TABLE `build_execution`
    MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `user`
--
ALTER TABLE `user`
    MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `user_action`
--
ALTER TABLE `user_action`
    MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT für Tabelle `user_variable`
--
ALTER TABLE `user_variable`
    MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- Constraints der exportierten Tabellen
--

--
-- Constraints der Tabelle `build_execution`
--
ALTER TABLE `build_execution`
    ADD CONSTRAINT `fk_build_definition_build_executions` FOREIGN KEY (`build_definition_id`) REFERENCES `build_definition` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
