--name: register-name
-- $1: primary_name
-- $2: blockchain_address
INSERT INTO alias(
    primary_name,
    blockchain_address
) VALUES($1, $2)

--name: update-name
-- $1: blockchain_address
-- $2: primary_name
UPDATE alias SET 
    blockchain_address = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE primary_name = $2 AND active = true

--name: lookup-name
-- $1: primary_name
SELECT blockchain_address FROM alias WHERE primary_name = $1 AND active = true

--name: reverse-lookup
-- $1: blockchain_address
SELECT primary_name FROM alias WHERE blockchain_address = $1 AND active = true