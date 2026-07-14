DROP INDEX contributions_campaign_raffle_number_key;
ALTER TABLE contributions DROP COLUMN raffle_number;
ALTER TABLE campaigns DROP COLUMN raffle_winning_number;
ALTER TABLE campaigns DROP COLUMN raffle_total_numbers;
ALTER TABLE campaigns DROP COLUMN raffle_unit_price;
