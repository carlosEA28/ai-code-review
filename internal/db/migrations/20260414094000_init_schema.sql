-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE review_job_status AS ENUM (
    'queued',
    'processing',
    'done',
    'failed'
);

CREATE TYPE comment_severity AS ENUM (
    'critical',
    'warning',
    'suggestion'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT UNIQUE NOT NULL,
    github_login TEXT NOT NULL,
    name TEXT,
    email CITEXT,
    avatar_url TEXT,
    github_token TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    github_repo_id BIGINT NOT NULL,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL,
    default_branch TEXT DEFAULT 'main',
    branch_filter TEXT,
    webhook_id BIGINT,
    webhook_secret TEXT NOT NULL,
    auto_review BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    last_review_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, github_repo_id)
);

CREATE TABLE prompt_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    system_prompt TEXT NOT NULL,
    model TEXT DEFAULT 'claude-sonnet-4-5',
    max_tokens INT DEFAULT 4096,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE pull_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    github_pr_id BIGINT NOT NULL,
    number INT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    author_login TEXT NOT NULL,
    base_branch TEXT NOT NULL,
    head_branch TEXT NOT NULL,
    head_sha TEXT NOT NULL,
    pr_url TEXT NOT NULL,
    diff_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (repository_id, github_pr_id)
);

CREATE TABLE review_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    prompt_config_id UUID REFERENCES prompt_configs(id),
    status review_job_status DEFAULT 'queued',
    model TEXT NOT NULL,
    input_tokens INT,
    output_tokens INT,
    duration_ms INT,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    queued_at TIMESTAMPTZ DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE TABLE review_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_job_id UUID NOT NULL REFERENCES review_jobs(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    line_number INT,
    diff_position INT,
    severity comment_severity,
    body TEXT NOT NULL,
    code_snippet TEXT,
    github_comment_id BIGINT,
    posted_to_github BOOLEAN DEFAULT false,
    posted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS review_comments;
DROP TABLE IF EXISTS review_jobs;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS prompt_configs;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS comment_severity;
DROP TYPE IF EXISTS review_job_status;

DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pgcrypto";
