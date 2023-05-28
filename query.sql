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
(id, created_at, updated_at, name, description, user_id, rule_id, score, moves, play_state, data, data_at_start, history, template_id, based_on_game, option_seed, option_state)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;
-- name: InsertRule :one
INSERT INTO rule
(id, slug, created_at, updated_at, description, mode, size_x, size_y, recreate_on_swipe, no_reswipe, no_multiply, no_addition, max_moves, target_cell_value, target_score, starting_cells)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;
-- name: InserTemplate :one
INSERT INTO game_template
(id, created_at, rule_id, created_by, name, description, challenge_number, ideal_moves, ideal_score, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

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

-- name: GetChallengeStatsForUser :many
SELECT
	g.id as game_id
	, g.template_id
	, g.score
	, g.moves
	, u.id as user_id
	, u.username 
FROM
	game AS g
	JOIN user AS u ON u.id = g.user_id 
WHERE
	template_id IS NOT NULL
  AND user_id = ?
	;
-- name: GetRule :one
select * from rule
where id == ? or slug == ?;
-- name: GetGame :one
select * from game
where id == ?;
-- name: GetOriginalGame :one
SELECT o.*
  FROM game AS g 
    JOIN game o ON o.id == g.based_on_game
 WHERE g.id = '?';
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
       game.option_seed game_option_seed,
       game.option_state game_option_state,
       game.data game_data,
       game.history game_history,
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
    data       = ?,
    history    = ?
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
     , (SELECT CAST(AVG(length(history)*length(history)) - AVG(length(history))*AVG(length(history)) as FLOAT) from game) as history_variance
     , (SELECT avg(length(history)) from game) as history_avg
     , (SELECT max(length(history)) from game) as history_max
     , (SELECT min(length(history)) from game) as history_min
     , (SELECT CAST(total(length(history)) as INT) from game where kind = 2) as history_total
