-- Payment and access records are the durable commercial records in Leamout.
-- A settled payment is not promoted into a generic commerce order.
DROP TABLE IF EXISTS orders;

COMMENT ON TABLE checkouts IS
    'Temporary billing sessions completed by payment settlement.';
