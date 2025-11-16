INSERT INTO team(id, team_name) 
VALUES 
    ('team1', 'Backend'),
    ('team2', 'Frontend'),
    ('team3', 'Analytics');

INSERT INTO team_member(id, username, activity, team_id)
VALUES
    ('u01', 'Edsger Dijkstra', 'ACTIVE', 'team1'),
    ('u02', 'Alan Turing', 'ACTIVE', 'team1'),
    ('u03', 'Robert Griesmer', 'ACTIVE', 'team1'),
    ('u04', 'Dennis Ritchie', 'ACTIVE', 'team1'),
    ('u05', 'Bjarne Stroustrup', 'INACTIVE', 'team1'),
    ('u06', 'Niklaus Wirth', 'ACTIVE', 'team1'),
    ('u07', 'Guido van Rossum', 'ACTIVE', 'team1'),

    ('u10', 'Vincent van Gogh', 'ACTIVE', 'team2'),
    ('u11', 'Ivan Aivazovsky', 'ACTIVE', 'team2'),
    ('u12', 'Alexandr Ivanov', 'ACTIVE', 'team2'),
    ('u13', 'Le Corbusier', 'ACTIVE', 'team2'),
    ('u14', 'Edvard Munch', 'INACTIVE', 'team2'),

    ('u20', 'Louis Armstrong', 'ACTIVE', 'team3'),
    ('u21', 'Duke Ellington', 'ACTIVE', 'team3'),
    ('u22', 'Frank Sinatra', 'INACTIVE', 'team3'),
    ('u23', 'Ella Fitzgerald', 'INACTIVE', 'team3');

INSERT INTO pull_request(id, pr_name, author_id, pr_status, team_id, created_at, merged_at)
VALUES
    ('pr1', 'Added user authentication middleware', 'u01', 'MERGED', 'team1', '2024-01-15 10:00:00', '2024-01-16 14:30:00'),
    ('pr2', 'Implemented database connection pool', 'u02', 'MERGED', 'team1', '2024-01-16 09:15:00', '2024-01-17 11:45:00'),
    ('pr3', 'Added idempotency check with Redis', 'u03', 'MERGED', 'team1', '2024-01-17 14:20:00', '2024-01-18 16:00:00'),
    ('pr4', 'Refactored application layer for better testing', 'u04', 'OPEN', 'team1', '2024-01-18 13:00:00', NULL),
    ('pr5', 'Added notification send to Kafka topic', 'u06', 'OPEN', 'team1', '2024-01-19 11:30:00', NULL),
    ('pr6', 'Added Elasticsearch for products', 'u07', 'OPEN', 'team1', '2024-01-19 16:45:00', NULL);

INSERT INTO assigned_reviewer(member_id, pr_id)
VALUES
    ('u02', 'pr1'),
    ('u03', 'pr1'),
    ('u04', 'pr2'),
    ('u07', 'pr2'),
    ('u01', 'pr3'),
    ('u06', 'pr3'),
    ('u02', 'pr4'),
    ('u07', 'pr4'),
    ('u03', 'pr5'),
    ('u04', 'pr5'),
    ('u01', 'pr6'),
    ('u02', 'pr6');