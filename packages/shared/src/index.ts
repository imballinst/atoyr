export type QTEResult =
  | { kind: 'miss' }
  | { kind: 'hit'; damage: number }
  | { kind: 'critical'; damage: number }
  | { kind: 'parry'; damage: number }
  | { kind: 'block'; damage: number }
  | { kind: 'none' };

export function pickRandomWord(words: string[]): string {
  if (words.length === 0) throw new Error('no words');
  return words[Math.floor(Math.random() * words.length)];
}

// returns damage and result kind
export function resolveQTE(
  attackerTime: number | null,
  defenderTime: number | null,
  baseDamage = 10,
): QTEResult {
  // attackerTime or defenderTime null means not completed
  if (attackerTime === null) return { kind: 'miss' };
  if (defenderTime === null) return { kind: 'critical', damage: baseDamage * 2 };

  const diff = attackerTime - defenderTime;
  const abs = Math.abs(diff);

  // way faster: difference less than -0.3s
  if (diff < -0.3) return { kind: 'critical', damage: Math.round(baseDamage * 1.75) };
  // faster: attacker faster than defender
  if (diff < 0) return { kind: 'hit', damage: Math.round(baseDamage * 1.0) };
  // slower but defender faster
  if (diff > 0 && diff < 0.3) return { kind: 'block', damage: Math.round(baseDamage * 0.25) };
  // way slower: defender parry
  if (diff >= 0.3) return { kind: 'parry', damage: Math.round(baseDamage * 1.0) };

  return { kind: 'none' };
}
