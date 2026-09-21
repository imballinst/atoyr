-- Only finished sessions belong to the immortalized "English words, July 2026"
-- leaderboard. Sessions that are still playing keep the "english-words" topic so
-- they can keep being assigned words from the new word list after the migration.
UPDATE session_entities
  SET topic = 'english-words-july-2026'
  WHERE topic = 'english-words' AND phase = 'finished';
