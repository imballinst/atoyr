import { Injectable } from '@nestjs/common';
import { resolveQTE, pickRandomWord } from '@atoyr/shared';

@Injectable()
export class MatchService {
  root() {
    return 'A Toy R server (NestJS prototype)';
  }

  matchmake() {
    const opponentWords = ['zeta', 'yummy', 'quick', 'apple', 'bravo'];
    return { opponentWords };
  }

  resolve(attackerTime?: number, defenderTime?: number) {
    const result = resolveQTE(
      typeof attackerTime === 'number' ? attackerTime : null,
      typeof defenderTime === 'number' ? defenderTime : null,
    );
    return { result };
  }
}
