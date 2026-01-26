export function validateAnswer(input: string, expected: string): boolean {
  return input.trim().toLowerCase() === expected.toLowerCase();
}
