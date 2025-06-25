-- Add unique constraint to blockchain_address to ensure one alias per user
ALTER TABLE alias ADD CONSTRAINT unique_blockchain_address UNIQUE (blockchain_address);
