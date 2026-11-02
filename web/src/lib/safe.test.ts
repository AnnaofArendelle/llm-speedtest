import { describe, expect, it } from 'vitest';
import { arr, entries, keys, obj, topEntries } from './safe';

describe('safe', () => {
  it('normalises null collections', () => {
    expect(arr(null)).toEqual([]);
    expect(arr(undefined)).toEqual([]);
    expect(arr([1, 2])).toEqual([1, 2]);
    expect(obj<number>(null)).toEqual({});
    expect(entries<number>(null)).toEqual([]);
    expect(keys(null)).toEqual([]);
    expect(entries({ a: 1 })).toEqual([['a', 1]]);
  });
  it('sorts top entries descending', () => {
    expect(topEntries({ a: 1, b: 3, c: 2 }, 2)).toEqual([
      ['b', 3],
      ['c', 2],
    ]);
    expect(topEntries(null)).toEqual([]);
  });
});
