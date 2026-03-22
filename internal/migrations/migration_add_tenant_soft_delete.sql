-- Migration: add multi-tenant columns, roles, business_members, audit_logs, and soft deletes
-- Do NOT run this automatically in this task (per instructions)

BEGIN;

-- Ensure businesses table exists before adding FK references
-- Add tenant and soft delete support to users
ALTER TABLE IF EXISTS users
  ADD COLUMN IF NOT EXISTS tenant_id BIGINT,
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL,
  ADD COLUMN IF NOT EXISTS role VARCHAR(20),
  ADD COLUMN IF NOT EXISTS status VARCHAR(20);

-- Add tenant foreign key (delay if businesses not present in DB at migration time)
-- This will succeed when businesses table exists; depending on migration ordering adjust as necessary
ALTER TABLE IF EXISTS users
  ADD CONSTRAINT IF NOT EXISTS fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES businesses(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Update businesses table: add plan, created_by, deleted_at
ALTER TABLE IF EXISTS businesses
  ADD COLUMN IF NOT EXISTS plan VARCHAR(20) DEFAULT 'free',
  ADD COLUMN IF NOT EXISTS created_by BIGINT,
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

ALTER TABLE IF EXISTS businesses
  ADD CONSTRAINT IF NOT EXISTS fk_business_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_businesses_deleted_at ON businesses(deleted_at);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY,
  business_id BIGINT NOT NULL,
  name VARCHAR(50) NOT NULL,
  permissions TEXT[] DEFAULT ARRAY[]::TEXT[],
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  CONSTRAINT fk_roles_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
  CONSTRAINT uq_roles_business_name UNIQUE (business_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roles_business ON roles(business_id);

-- Create business_members table
CREATE TABLE IF NOT EXISTS business_members (
  id BIGSERIAL PRIMARY KEY,
  business_id BIGINT NOT NULL,
  user_id BIGINT,
  email VARCHAR(255),
  role_id BIGINT NOT NULL,
  status VARCHAR(20) DEFAULT 'pending',
  invited_by BIGINT,
  invited_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  accepted_at TIMESTAMP WITH TIME ZONE NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  CONSTRAINT fk_bm_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
  CONSTRAINT fk_bm_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  CONSTRAINT fk_bm_role FOREIGN KEY (role_id) REFERENCES roles(id),
  CONSTRAINT fk_bm_invited_by FOREIGN KEY (invited_by) REFERENCES users(id),
  CONSTRAINT uq_business_member_email UNIQUE (business_id, email)
);

CREATE INDEX IF NOT EXISTS idx_business_members_business ON business_members(business_id);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  business_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  action VARCHAR(100) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_id BIGINT,
  old_values JSONB,
  new_values JSONB,
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  CONSTRAINT fk_audit_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
  CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_business_time ON audit_logs(business_id, created_at DESC);

-- Add indexes to speed up tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_users_tenant_deleted ON users(tenant_id, deleted_at);

COMMIT;
