-- name: InsertUser :one
INSERT INTO user
    (id, created_at, updated_at, username, active_game_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;
-- name: InsertSession :one
INSERT INTO session
    (id, created_at, updated_at, invalid_after, user_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;
-- name: InsertGame :one
INSERT INTO game
(id, created_at, updated_at, name, description, user_id, rule_id, score, moves, play_state, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;
-- name: InsertRule :one
INSERT INTO rule
(id, slug, created_at, updated_at, description, mode, size_x, size_y, recreate_on_swipe, no_reswipe, no_multiply, no_addition, max_moves, target_cell_value, target_score)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;
-- name: InserTemplate :one
INSERT INTO game_template
(id, created_at, rule_id, created_by, name, description, challenge_number, ideal_moves, ideal_score, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: InsertGameHistory :one
insert into game_history
(created_at, game_id, move, kind, points, data) 
values 
(?, ?, ?, ?, ?, ?)
RETURNING *;
-- name: GetGameHistoryByMoveNumber :one
select * from game_history
where game_id = ? and move == ?;
-- name: GetGameTemplate :one
select * from game_template
where id = ?;
-- name: GetGameTemplateByChallengeNumber :one
select * from game_template
where challenge_number = ?;
-- name: GetGameChallengesTemplates :many
select * from game_template
where challenge_number is not null
order by challenge_number;
-- name: GetRule :one
select * from rule
where id == ? or slug == ?;
-- name: GetGame :one
select * from game
where id == ?;
-- name: GetUser :one
select * from user
where id == ?;
-- name: GetUserBySessionID :one
SELECT
       session.id session_id,
       session.created_at session_created_at,
       session.invalid_after session_invalid_after,

       user.id user_id,
       user.created_at user_created_at,
       user.updated_at user_updated_at,
       user.username,

       game.id game_id,
       game.created_at game_created_at,
       game.updated_at game_updated_at,
       game.description game_description,
       game.name game_Name,
       game.data game_data,
       game.play_state game_play_state,
       game.score game_score,
       game.moves game_moves,
       game.rule_id rule_id

FROM session session
         INNER JOIN user user on user.id = session.user_id
         INNER JOIN game game on user.active_game_id = game.id
WHERE session.id = ? LIMIT 1;

-- name: GetAllGames :many
SELECT * from game;
-- name: GetAllGameHistory :many
SELECT * from game_history;
-- name: GetAllRules :many
SELECT * from rule;
-- name: GetAllSessions :many
SELECT * from session;
-- name: GetAllUsers :many
SELECT * from user;
-- name: GetAllTemplates :many
SELECT * from game_template;
-- name: UpdateGame :one
UPDATE game
SET updated_at = ?,
    user_id    = ?,
    rule_id    = ?,
    score      = ?,
    moves      = ?,
    play_state = ?,
    data       = ?
WHERE id = ?
RETURNING *;
-- name: SetPlayStateForGame :one
UPDATE game
SET updated_at = ?,
    play_state = ?
WHERE id = ?
RETURNING *;
-- name: UpdateUser :one
UPDATE user
SET updated_at = ?,
    username = ?,
    active_game_id = ?
WHERE id = ?
RETURNING *;
-- name: SetActiveGameFormUser :one
UPDATE user
SET updated_at = ?,
    active_game_id = ?
WHERE id = ?
RETURNING *;
-- name: Stats :one
SELECT (SELECT COUNT(*) FROM user) AS users
     , (SELECT COUNT(*) FROM session) AS session
     , (SELECT COUNT(*) FROM game) AS games
     , (SELECT COUNT(*) FROM game where game.play_state = 1) AS games_won
     , (SELECT COUNT(*) FROM game where game.play_state = 2) AS games_lost
     , (SELECT COUNT(*) FROM game where game.play_state = 3) AS games_abandoned
     , (SELECT COUNT(*) FROM game where game.play_state = 4) AS games_current
     , (SELECT max(game.moves) FROM game where game.play_state = 4) AS longest_game
     , (SELECT max(game.score) FROM game where game.play_state = 4) AS highest_score
     , (SELECT CAST(AVG(length(data)*length(data)) - AVG(length(data))*AVG(length(data)) as FLOAT) from game_history where kind = 2) as history_data_variance
     , (SELECT avg(length(data)) from game_history where kind = 2) as combine_data_avg
     , (SELECT max(length(data)) from game_history where kind = 2) as combine_data_max
     , (SELECT min(length(data)) from game_history where kind = 2) as combine_data_min
     , (SELECT CAST(total(length(data)) as INT) from game_history where kind = 2) as combine_data_total
