-- MechBot MySQL 建表脚本（需要 MySQL 8.0+），与 internal/store 里的 SQLite 表结构一一对应
-- 执行: mysql -u root -p < schema/mechbot.sql
-- 可重复执行，已存在的表不会被改动。
-- 目前后端仍用 SQLite；切换到 MySQL 时还要改 store 层的 SQL 方言（INSERT OR IGNORE、ON CONFLICT 等）。

CREATE DATABASE IF NOT EXISTS mechbot DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
USE mechbot;

-- 用户。expire_at 为 RFC3339 文本，空表示从未开通 Pro
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

-- 用户监控的钱包地址
CREATE TABLE IF NOT EXISTS addresses (
  address           VARCHAR(64) NOT NULL,
  telegram_id       BIGINT      NOT NULL,
  min_health_factor DOUBLE      NOT NULL,
  created_at        VARCHAR(40) NOT NULL,
  PRIMARY KEY (address, telegram_id),
  KEY addresses_user (telegram_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 每个地址每个协议的预警状态，用于去重和冷却
CREATE TABLE IF NOT EXISTS alert_state (
  address         VARCHAR(64) NOT NULL,
  telegram_id     BIGINT      NOT NULL,
  protocol_key    VARCHAR(64) NOT NULL,
  in_danger       TINYINT     NOT NULL DEFAULT 0,
  last_alert_unix BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (address, telegram_id, protocol_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 最近一次查询到的仓位
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

-- 网页登录：Telegram 里确认后换成 session
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

-- 每个用户的专属充值地址。idx 是 HD 派生序号，从 1 开始，0 留给归集用的 gas 钱包
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

-- 链上到账记录。主键保证同一笔转账只入账一次；days = 0 表示金额不够任何套餐
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

-- 各 EVM 链已扫描到的区块
CREATE TABLE IF NOT EXISTS scan_cursors (
  chain VARCHAR(16)     NOT NULL,
  block BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (chain)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 推荐奖励，每个被推荐人只奖励一次
CREATE TABLE IF NOT EXISTS referral_rewards (
  referee_id   BIGINT NOT NULL,
  referrer_id  BIGINT NOT NULL,
  days         INT    NOT NULL,
  created_unix BIGINT NOT NULL,
  PRIMARY KEY (referee_id),
  KEY referral_rewards_referrer (referrer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
