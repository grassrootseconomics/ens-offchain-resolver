--name: register-name
-- $1: primary_name
-- $2: blockchain_address
INSERT INTO alias(
    primary_name,
    blockchain_address
) VALUES($1, $2)

--name: update-name
-- $1: primary_name
-- $2: blockchain-address
UPDATE alias SET 
    primary_name = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE blockchain_address = $2 AND active = true

--name: lookup-name
-- $1: primary_name
SELECT blockchain_address FROM alias WHERE primary_name = $1 AND active = true

--name: reverse-lookup
-- $1: blockchain_address
SELECT primary_name FROM alias WHERE blockchain_address = $1 AND active = true

--name: upsert-name
-- $1: primary_name
-- $2: blockchain_address
INSERT INTO alias(primary_name, blockchain_address) 
VALUES($1, $2)
ON CONFLICT (blockchain_address) 
DO UPDATE SET 
    primary_name = EXCLUDED.primary_name,
    updated_at = CURRENT_TIMESTAMP