CREATE DATABASE IF NOT EXISTS helgart;

USE helgart;

CREATE TABLE IF NOT EXISTS products (
    id bigint(20) not null AUTO_INCREMENT,
    exchange varchar(50) not null,
    pair varchar(10) not null,
    ex_base varchar(10) not null,
    ex_quote varchar(10) not null,
    he_base varchar(10) not null,
    he_quote varchar(10) not null,
    PRIMARY KEY(id)
);

SELECT 
    pair, 
    COUNT(pair)
FROM
    products
GROUP BY pair
HAVING COUNT(pair) > 1
LIMIT 10;

-- PAIRS WITH MORE THAN ONE MARKET
-- SELECT count(*) FROM (SELECT      pair,      COUNT(pair) FROM     products GROUP BY pair HAVING COUNT(pair) = 1) as data;