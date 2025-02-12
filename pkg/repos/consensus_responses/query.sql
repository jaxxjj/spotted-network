-- name: CreateConsensusResponse :one
-- -- timeout: 500ms
INSERT INTO consensus_responses (
    task_id,
    epoch,
    value,
    key,
    chain_id, 
    block_number,
    target_address,
    aggregated_signatures,
    operator_addresses
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetConsensusResponseByTaskId :one
-- -- timeout: 500ms
SELECT * FROM consensus_responses
WHERE task_id = $1 LIMIT 1;

-- name: GetConsensusResponseByRequest :one
-- -- timeout: 500ms
SELECT * FROM consensus_responses
WHERE target_address = $1 
AND chain_id = $2 
AND block_number = $3 
AND key = $4 
LIMIT 1;

