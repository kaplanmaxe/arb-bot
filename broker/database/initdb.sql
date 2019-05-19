CREATE DATABASE IF NOT EXISTS helgart;

USE helgart;

CREATE TABLE IF NOT EXISTS products (
    id bigint(20) not null AUTO_INCREMENT,
    exchange varchar(50) not null,
    ex_pair varchar(20) not null,
    he_pair varchar(20) not null,
    ex_base varchar(10) not null,
    ex_quote varchar(10) not null,
    he_base varchar(10) not null,
    he_quote varchar(10) not null,
    PRIMARY KEY(id)
);

-- SELECT 
--     ex_pair, 
--     COUNT(ex_pair)
-- FROM
--     products
-- GROUP BY ex_pair
-- HAVING COUNT(ex_pair) > 1
-- LIMIT 10;

-- PAIRS WITH MORE THAN ONE MARKET
-- SELECT count(*) FROM (SELECT      pair,      COUNT(pair) FROM     products GROUP BY pair HAVING COUNT(pair) = 1) as data;