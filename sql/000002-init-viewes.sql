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

CREATE VIEW assignments_per_members AS
SELECT m.id AS member_id, COUNT(a.pr_id) AS assigments_count
FROM team_member AS m
LEFT JOIN assigned_reviewer AS a
    ON a.member_id = m.id
GROUP BY m.id
ORDER BY assigments_count DESC;