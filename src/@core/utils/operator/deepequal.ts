import { deepStrictEqual } from 'assert';

/**
 * deepEqual checks if two given params are completely equal (not only their value, but also types), and returns true if so.
 * 
 * Documentation:
 * https://nodejs.org/api/assert.html#assert_assert_deepstrictequal_actual_expected_message
 *
 * @export
 * @param {*} a The first param to compare with.
 * @param {*} b The second param to compare with.
 * @returns {boolean} The compared result.
 */
export function deepEqual(a: any, b: any): boolean {
    try {
        deepStrictEqual(a, b)
    } catch (e) {
        return false;
    }
    return true;
}