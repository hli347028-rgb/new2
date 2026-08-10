-- AIX schema (from docs/AIX数据库设计.md §8) + seed data
CREATE DATABASE IF NOT EXISTS aix DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE aix;

CREATE TABLE IF NOT EXISTS users (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  address           VARCHAR(42)     NOT NULL,
  inviter_id        BIGINT UNSIGNED NULL,
  invite_code       VARCHAR(64)     NOT NULL,
  usdt_recharge     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  usdt_reward       DECIMAL(36,18)  NOT NULL DEFAULT 0,
  aix_balance       DECIMAL(36,18)  NOT NULL DEFAULT 0,
  mgmt_level        TINYINT UNSIGNED NOT NULL DEFAULT 0,
	large_area_perf  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  small_area_perf   DECIMAL(36,18)  NOT NULL DEFAULT 0,
  team_perf         DECIMAL(36,18)  NOT NULL DEFAULT 0,
  status            TINYINT         NOT NULL DEFAULT 1,
  created_time      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_address (address),
  UNIQUE KEY uk_users_invite_code (invite_code),
  KEY idx_users_inviter_id (inviter_id),
  KEY idx_users_mgmt_level (mgmt_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS orders (
  id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id        BIGINT UNSIGNED NOT NULL,
  principal      DECIMAL(36,18)  NOT NULL,
  exit_cap       DECIMAL(36,18)  NOT NULL,
  earned_total   DECIMAL(36,18)  NOT NULL DEFAULT 0,
  direct_base    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_recharge  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_reward    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  fund_source    VARCHAR(16)     NOT NULL,
  status         VARCHAR(16)     NOT NULL DEFAULT 'active',
  exited_time    DATETIME(3)     NULL,
  created_time   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_orders_user_status (user_id, status),
  KEY idx_orders_status_created (status, created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS recharges (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id          BIGINT UNSIGNED NOT NULL,
  amount           DECIMAL(36,18)  NOT NULL,
  tx_hash          VARCHAR(66)     NOT NULL,
  from_address     VARCHAR(42)     NULL,
  to_address       VARCHAR(42)     NULL,
  status           VARCHAR(16)     NOT NULL DEFAULT 'pending',
  confirmed_time   DATETIME(3)     NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_recharges_tx_hash (tx_hash),
  KEY idx_recharges_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS transfers (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  from_user_id         BIGINT UNSIGNED NOT NULL,
  to_user_id           BIGINT UNSIGNED NOT NULL,
  asset                VARCHAR(16)     NOT NULL,
  amount               DECIMAL(36,18)  NOT NULL,
  pay_from             VARCHAR(16)     NULL,
  from_recharge_debit  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_reward_debit    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  to_credit_reward     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  to_credit_aix        DECIMAL(36,18)  NOT NULL DEFAULT 0,
  remark               VARCHAR(255)    NULL,
  created_time         DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_transfers_from_created (from_user_id, created_time),
  KEY idx_transfers_to_created (to_user_id, created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS reward_logs (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id          BIGINT UNSIGNED NOT NULL,
  from_user_id     BIGINT UNSIGNED NULL,
  order_id         BIGINT UNSIGNED NULL,
  batch_id         BIGINT UNSIGNED NULL,
  type             VARCHAR(32)     NOT NULL,
  asset            VARCHAR(16)     NOT NULL,
  amount           DECIMAL(36,18)  NOT NULL,
  base_amount      DECIMAL(36,18)  NULL,
  rate             DECIMAL(36,18)  NULL,
  exit_applied     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  meta             JSON            NULL,
  settlement_date  DATE            NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_reward_user_type_created (user_id, type, created_time),
  KEY idx_reward_batch (batch_id),
  KEY idx_reward_order (order_id),
  KEY idx_reward_settlement_date (settlement_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS aix_prices (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  price           DECIMAL(36,18)  NOT NULL,
  effective_date  DATE            NOT NULL,
  remark          VARCHAR(255)    NULL,
  created_time    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_aix_prices_effective_date (effective_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS settlement_batches (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  settlement_date  DATE            NOT NULL,
  aix_price        DECIMAL(36,18)  NOT NULL,
  status           VARCHAR(16)     NOT NULL DEFAULT 'running',
  static_count     INT UNSIGNED    NOT NULL DEFAULT 0,
  static_amount    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  mgmt_count       INT UNSIGNED    NOT NULL DEFAULT 0,
  mgmt_amount      DECIMAL(36,18)  NOT NULL DEFAULT 0,
  started_time     DATETIME(3)     NULL,
  finished_time    DATETIME(3)     NULL,
  error_msg        VARCHAR(512)    NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_settlement_date (settlement_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS settings (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `key`         VARCHAR(64)     NOT NULL,
  value         JSON            NOT NULL,
  created_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_settings_key (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed system_config
INSERT INTO settings (`key`, value)
VALUES (
  'system_config',
  JSON_OBJECT(
    'static_rate', 0.5,
    'exit_multiplier', 4,
    'direct_rate', 0.5,
    'mgmt_thresholds', JSON_ARRAY(5000, 20000, 50000, 200000, 500000, 1500000, 4000000, 8000000, 15000000, 30000000),
    'mgmt_rates', JSON_ARRAY(0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1),
    'aix_price_initial', 1,
    'mgmt_counts_toward_exit', true,
    'min_subscribe', '100'
  )
)
ON DUPLICATE KEY UPDATE value = VALUES(value);

-- Seed today's AIX price = 1
INSERT INTO aix_prices (price, effective_date, remark)
VALUES (1, CURDATE(), 'initial')
ON DUPLICATE KEY UPDATE price = VALUES(price);
