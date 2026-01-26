import { WordEntry } from '@atoyr/shared';
import { Injectable } from '@nestjs/common';
import wordData from '../data/words.json';

@Injectable()
export class WordService {
  private words: WordEntry[] = wordData as WordEntry[];

  /**
   * Get a random word that hasn't been used in the session
   * @param usedWords Array of words already used
   * @returns A random word or null if all words have been used
   */
  async getRandomWord(usedWords: string[]): Promise<WordEntry | null> {
    const available = this.words.filter((w) => !usedWords.includes(w.word));

    if (available.length === 0) return null;

    return available[Math.floor(Math.random() * available.length)];
  }

  /**
   * Get all words
   */
  getAllWords(): WordEntry[] {
    return this.words;
  }

  /**
   * Get total word count
   */
  getWordCount(): number {
    return this.words.length;
  }
}
