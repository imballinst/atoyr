## Tech spec

- Use AccelByte Gaming Services and AccelByte Multiplayer Servers
- Use TypeScript
- Use Prettier for formatting
- Use Yarn Modern with nodeLinker node_modules
- Use any kind of graphical language
- No authentication is required (yet)

## Game mechanics

- Turn based game
- Multiplayer, 1v1 matchmaking
- A player can level up, increasing stats and all that
- When a player gets matchmade with another with significantly more level and stats, the latter will be "synchronized" to not get significant advantage. The other player will receive bonus rewards if he beats them
- There will be a QTE. Each match, a player will pick 5 predefined words. Each turn, each player will randomly get picked a word that their opponent picked. They will do QTE, the player with faster time to completion will land more damage to their enemy
- If the attacker's QTE is way faster than the defender, then the attacker will do critical damage
- If the attacker's QTE is faster than the defender, then the attacker will do damage
- If the attacker's QTE is slower than the defender, then the attacker's damage will be partially blocked
- If the attacker's QTE is not completed, then the attacker's attack will miss
- If the defender's QTE is way faster than the attacker, then the defender will parry and counter with base damage
- If the defender's QTE is faster than the attacker, then the defender will partially block the damage
- If the defender's QTE is slower than the attacker, then the defender will get damaged
- If the defender's QTE is not completed, then the defender will get critically damaged
- If equal, then no damage will happen
