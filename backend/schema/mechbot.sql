-- MechBot MySQL schema (MySQL 8.0+), mirroring the SQLite tables in internal/store
-- Run: mysql -u root -p < schema/mechbot.sql
-- Safe to rerun; existing tables are left untouched.
-- The backend still uses SQLite; moving to MySQL also means porting the store layer's SQL dialect (INSERT OR IGNORE, ON CONFLICT, ...).

CREATE DATABASE IF NOT EXISTS mechbot DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
USE mechbot;

-- Users. expire_at is RFC3339 text; empty means Pro was never activated
CREATE TABLE IF NOT EXISTS users (
  telegram_id       BIGINT       NOT NULL,
  username          VARCHAR(64)  NOT NULL DEFAULT '',
  first_name        VARCHAR(128) NOT NULL DEFAULT '',
  ref_by            BIGINT       NULL,
  expire_at         VARCHAR(40)  NULL,
  is_premium        TINYINT      NOT NULL DEFAULT 0,
  min_health_factor DOUBLE       NOT NULL DEFAULT 1.2,
  created_at        VARCHAR(40)  NOT NULL,
  PRIMARY KEY (telegram_id),
  KEY users_ref_by (ref_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Wallet addresses each user watches
CREATE TABLE IF NOT EXISTS addresses (
  address           VARCHAR(64) NOT NULL,
  telegram_id       BIGINT      NOT NULL,
  min_health_factor DOUBLE      NOT NULL,
  created_at        VARCHAR(40) NOT NULL,
  PRIMARY KEY (address, telegram_id),
  KEY addresses_user (telegram_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Alert state per address and protocol, for dedup and cooldown
CREATE TABLE IF NOT EXISTS alert_state (
  address         VARCHAR(64) NOT NULL,
  telegram_id     BIGINT      NOT NULL,
  protocol_key    VARCHAR(64) NOT NULL,
  in_danger       TINYINT     NOT NULL DEFAULT 0,
  last_alert_unix BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (address, telegram_id, protocol_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Latest position reading
CREATE TABLE IF NOT EXISTS snapshots (
  address        VARCHAR(64)  NOT NULL,
  protocol_key   VARCHAR(64)  NOT NULL,
  label          VARCHAR(128) NOT NULL,
  health_factor  DOUBLE       NOT NULL DEFAULT 0,
  collateral_usd DOUBLE       NOT NULL DEFAULT 0,
  debt_usd       DOUBLE       NOT NULL DEFAULT 0,
  error          TEXT         NOT NULL,
  PRIMARY KEY (address, protocol_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS address_checks (
  address      VARCHAR(64) NOT NULL,
  checked_unix BIGINT      NOT NULL,
  PRIMARY KEY (address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Web login: exchanged for a session once confirmed in Telegram
CREATE TABLE IF NOT EXISTS login_tokens (
  token        VARCHAR(64) NOT NULL,
  code         VARCHAR(16) NOT NULL,
  ref_by       BIGINT      NULL,
  telegram_id  BIGINT      NULL,
  status       VARCHAR(16) NOT NULL DEFAULT 'pending',
  created_unix BIGINT      NOT NULL,
  PRIMARY KEY (token),
  KEY login_tokens_created (created_unix)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS sessions (
  id           VARCHAR(128) NOT NULL,
  telegram_id  BIGINT       NOT NULL,
  expires_unix BIGINT       NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Each user's personal deposit address. idx is the HD derivation index, starting at 1; 0 is reserved for the sweep gas wallet
CREATE TABLE IF NOT EXISTS deposit_addresses (
  telegram_id      BIGINT       NOT NULL,
  idx              INT UNSIGNED NOT NULL,
  evm              VARCHAR(42)  NOT NULL DEFAULT '',
  tron             VARCHAR(34)  NOT NULL DEFAULT '',
  watch_until_unix BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (telegram_id),
  UNIQUE KEY deposit_addresses_idx (idx),
  KEY deposit_addresses_evm (evm),
  KEY deposit_addresses_tron (tron)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- On-chain payments. The primary key credits each transfer only once; days = 0 means the amount matched no plan
CREATE TABLE IF NOT EXISTS deposits (
  chain        VARCHAR(16) NOT NULL,
  tx_hash      VARCHAR(80) NOT NULL,
  log_index    INT         NOT NULL,
  telegram_id  BIGINT      NOT NULL,
  amount_micro BIGINT      NOT NULL,
  days         INT         NOT NULL,
  status       VARCHAR(16) NOT NULL,
  created_unix BIGINT      NOT NULL,
  PRIMARY KEY (chain, tx_hash, log_index),
  KEY deposits_user (telegram_id, created_unix)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Last scanned block per EVM chain
CREATE TABLE IF NOT EXISTS scan_cursors (
  chain VARCHAR(16)     NOT NULL,
  block BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (chain)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Referral rewards, at most one per referred user
CREATE TABLE IF NOT EXISTS referral_rewards (
  referee_id   BIGINT NOT NULL,
  referrer_id  BIGINT NOT NULL,
  days         INT    NOT NULL,
  created_unix BIGINT NOT NULL,
  PRIMARY KEY (referee_id),
  KEY referral_rewards_referrer (referrer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
