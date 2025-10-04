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
  attackerText?: string | null,
  defenderText?: string | null,
  expectedWord?: string | null,
  baseDamage = 10,
): QTEResult {
  const makeDamage = (kind: Exclude<QTEResult['kind'], 'miss' | 'none'>, raw: number) => {
    const dmg = Math.round(raw);
    return { kind, damage: Math.max(1, dmg) } as any;
  };
  // If no text args are provided, preserve the original timing-only behavior
  const textArgsProvided =
    typeof attackerText !== 'undefined' ||
    typeof defenderText !== 'undefined' ||
    typeof expectedWord !== 'undefined';
  if (!textArgsProvided) {
    // attackerTime or defenderTime null means not completed
    if (attackerTime === null) return { kind: 'miss' };
    // defender did not complete -> treat as critical at 1.5x base damage
    if (defenderTime === null) return makeDamage('critical', baseDamage * 1.5);

    const diff = attackerTime - defenderTime;
    const abs = Math.abs(diff);

    // way faster: difference less than -0.3s -> critical (1.5x)
    if (diff < -0.3) return makeDamage('critical', baseDamage * 1.5);
    // faster: attacker faster than defender
    if (diff < 0) return makeDamage('hit', baseDamage * 1.0);
    // slower but defender faster
    if (diff > 0 && diff < 0.3) return { kind: 'block', damage: Math.round(baseDamage * 0.25) };
    // way slower: defender parry
    if (diff >= 0.3) return { kind: 'parry', damage: Math.round(baseDamage * 1.0) };

    return { kind: 'none' };
  }
  // helper: levenshtein distance
  function levenshtein(a: string, b: string) {
    if (a === b) return 0;
    const m = a.length;
    const n = b.length;
    const dp: number[][] = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0));
    for (let i = 0; i <= m; i++) dp[i][0] = i;
    for (let j = 0; j <= n; j++) dp[0][j] = j;
    for (let i = 1; i <= m; i++) {
      for (let j = 1; j <= n; j++) {
        const cost = a[i - 1] === b[j - 1] ? 0 : 1;
        dp[i][j] = Math.min(dp[i - 1][j] + 1, dp[i][j - 1] + 1, dp[i - 1][j - 1] + cost);
      }
    }
    return dp[m][n];
  }

  const word = typeof expectedWord === 'string' ? expectedWord : '';

  const calcAccuracy = (text?: string | null) => {
    if (!text || word.length === 0) return 0;
    const dist = levenshtein(text.toLowerCase(), word.toLowerCase());
    return Math.max(0, 1 - dist / Math.max(1, word.length));
  };

  const attackerAcc = calcAccuracy(attackerText ?? null);
  const defenderAcc = calcAccuracy(defenderText ?? null);

  // attacker missing text -> miss
  if (!attackerText) return { kind: 'miss' };
  // defender missing -> attacker critical
  if (!defenderText)
    // defender missing -> critical at 1.5x base damage (do not scale further with accuracy)
    return makeDamage('critical', baseDamage * 1.5);

  const diff = attackerTime !== null && defenderTime !== null ? attackerTime - defenderTime : 0;

  // Accuracy-first rules:
  // 1) If defender is significantly more accurate and faster, parry.
  if (defenderAcc > attackerAcc + 0.2 && diff >= 0.25) {
    return makeDamage('parry', baseDamage * (0.5 + defenderAcc));
  }

  // compute a time factor from relative submission times: attackerTime proportion of total
  // smaller attackerTime means more time remaining -> larger factor. Clamp to [0.1, 1].
  let timeFactor = 1;
  if (attackerTime !== null && defenderTime !== null && attackerTime + defenderTime > 0) {
    timeFactor = 1 - attackerTime / (attackerTime + defenderTime);
    timeFactor = Math.max(0.1, Math.min(1, timeFactor));
  }

  const ACC_CRIT_DELTA = 0.25; // if attacker accuracy exceeds defender by this much -> critical

  // If attacker is substantially more accurate -> attacker wins by accuracy
  if (attackerAcc > defenderAcc + 0.02) {
    // critical if accuracy advantage is large
    if (attackerAcc - defenderAcc >= ACC_CRIT_DELTA || diff < -0.3) {
      // critical scaled by accuracy and time remaining
      const raw = baseDamage * 1.5 * (0.8 + attackerAcc) * timeFactor;
      return makeDamage('critical', raw);
    }
    // normal hit: scale by accuracy and time remaining
    const raw = baseDamage * (0.8 + attackerAcc) * timeFactor;
    return makeDamage('hit', raw);
  }

  // If accuracies are very close, fall back to timing-driven outcomes but still apply timeFactor
  // way faster -> critical
  if (diff < -0.3) return makeDamage('critical', baseDamage * 1.5 * timeFactor);
  // faster -> hit
  if (diff < 0) return makeDamage('hit', baseDamage * (0.8 + attackerAcc) * timeFactor);
  // slightly slower -> block
  if (diff > 0 && diff < 0.3)
    return makeDamage('block', baseDamage * 0.25 * (0.5 + attackerAcc) * timeFactor);
  // way slower -> parry
  if (diff >= 0.3) return makeDamage('parry', baseDamage * (0.5 + defenderAcc) * timeFactor);

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
  // Defensive coercion: ensure numeric values to avoid NaN if fields were corrupted
  const dmg = Number(rawDamage) || 0;
  const defense = Number(target.defense) || 0;
  const currentHp = Number(target.hp) || 0;
  const mitigated = Math.max(0, Math.round(dmg - defense));
  target.hp = Math.max(0, currentHp - mitigated);
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
