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

// --- Player & Match helpers ---
export type Player = {
  id: string;
  name?: string;
  level: number;
  maxHp: number;
  hp: number;
  attack: number;
  defense: number;
  words: string[];
};

export function statsForLevel(level: number) {
  const maxHp = 100 + level * 10;
  const attack = 10 + level * 2;
  const defense = 5 + Math.floor(level * 1.2);
  return { maxHp, attack, defense };
}

export function createPlayer(
  id: string,
  name: string | undefined,
  level = 1,
  words: string[] = [],
): Player {
  const s = statsForLevel(level);
  return {
    id,
    name,
    level,
    maxHp: s.maxHp,
    hp: s.maxHp,
    attack: s.attack,
    defense: s.defense,
    words,
  };
}

export function synchronizePlayers(weak: Player, strong: Player) {
  // If strong is much higher level than weak, scale down strong's attack/defense
  const levelDiff = strong.level - weak.level;
  if (levelDiff <= 3) return strong; // no sync needed
  const factor = 1 - Math.min(0.5, (levelDiff - 3) * 0.1); // reduce up to 50%
  return {
    ...strong,
    attack: Math.max(1, Math.round(strong.attack * factor)),
    defense: Math.max(0, Math.round(strong.defense * factor)),
  };
}

export function applyDamage(target: Player, rawDamage: number) {
  const mitigated = Math.max(0, Math.round(rawDamage - target.defense));
  target.hp = Math.max(0, target.hp - mitigated);
  return mitigated;
}

export function levelUpIfEligible(player: Player, xp: number) {
  // simple rule: xp >= 100 -> +1 level, reduce xp
  // placeholder: not tracking xp persistently here
  if (xp >= 100) {
    player.level += Math.floor(xp / 100);
    const s = statsForLevel(player.level);
    player.maxHp = s.maxHp;
    player.attack = s.attack;
    player.defense = s.defense;
    player.hp = player.maxHp;
  }
}
