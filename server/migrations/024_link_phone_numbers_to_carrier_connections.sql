ALTER TABLE phone_numbers
    ADD COLUMN carrier_connection_id UUID,
    ADD COLUMN provider_resource_id TEXT;

ALTER TABLE phone_numbers
    ADD CONSTRAINT fk_phone_numbers_carrier_connection
        FOREIGN KEY (carrier_connection_id)
        REFERENCES carrier_connections(id)
        ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_phone_numbers_carrier_connection_id
    ON phone_numbers (carrier_connection_id)
    WHERE carrier_connection_id IS NOT NULL;

ALTER TABLE phone_numbers
    DROP COLUMN provider_id;
