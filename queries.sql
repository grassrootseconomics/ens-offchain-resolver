--name: register-name
-- $1: primary_name
-- $2: blockchain_address
INSERT INTO alias(
    primary_name,
    blockchain_address
) VALUES($1, $2)

--name: lookup-name
-- $1: primary_name
SELECT blockchain_address FROM alias WHERE primary_name = $1 AND active = true

--name: reverse-lookup
-- $1: blockchain_address
SELECT primary_name FROM alias WHERE blockchain_address = $1 AND active = true