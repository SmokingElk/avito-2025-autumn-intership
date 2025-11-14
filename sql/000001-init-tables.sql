CREATE TABLE IF NOT EXISTS team (
    -- add team id for possible team renaming
    id        VARCHAR(36) PRIMARY KEY,
    team_name VARCHAR(64) UNIQUE NOT NULL
);

-- add index on team_name, because current api provides team search by name
CREATE INDEX IF NOT EXISTS idx_team_name ON team(team_name);

CREATE TABLE IF NOT EXISTS team_member (
    -- it is possible to migrate uuids of users from other service
    -- in my opinion it is potentialy potentially better to replace string to UUID
    id       VARCHAR(36) PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    -- admit possible extension of states set
    activity VARCHAR(16) NOT NULL,
    team_id  VARCHAR(64) REFERENCES team(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS pull_request (
    id         VARCHAR(36) PRIMARY KEY,
    pr_name    VARCHAR(128) NOT NULL,
    author_id  VARCHAR(36) REFERENCES team_member(id) ON DELETE NO ACTION NOT NULL,
    pr_status  VARCHAR(16) NOT NULL,
    team_id    VARCHAR(36) REFERENCES team(id) ON DELETE NO ACTION NOT NULL,

    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    merged_at  TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE IF NOT EXISTS assigned_reviewer (
    member_id VARCHAR(36) REFERENCES team_member(id) ON DELETE CASCADE,
    pr_id     VARCHAR(36) REFERENCES pull_request(id) ON DELETE CASCADE,

    PRIMARY KEY (member_id, pr_id)
);

CREATE VIEW pr_with_members AS 
SELECT
    pr.id,
    pr.pr_name,
    pr.author_id,
    pr.pr_status,
    pr.created_at,
    pr.merged_at,
    pr.team_id,
    ARRAY_AGG(a.member_id) FILTER (WHERE a.member_id IS NOT NULL) AS reviewers
FROM pull_request AS pr
LEFT JOIN assigned_reviewer AS a
    ON pr.id = a.pr_id
GROUP BY pr.id;