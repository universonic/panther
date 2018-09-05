/**
 * deepCopy dives into given `obj`, copy each of its properties, and returns a duplicated new instance.
 *
 * @export
 * @param {*} obj The given source to deep copy from.
 * @returns The duplicated instance.
 */
export function deepCopy(obj: any): any {
    let buf;

    // Handle the 3 simple types, and null or undefined
    if (obj == null || "object" != typeof obj) return obj;

    // Handle Date
    if (obj instanceof Date) {
        buf = new Date();
        buf.setTime(obj.getTime());
        return buf;
    }

    // Handle Array
    if (obj instanceof Array) {
        buf = [];
        for (var i = 0, len = obj.length; i < len; i++) {
            buf[i] = deepCopy(obj[i]);
        }
        return buf;
    }

    // Handle Object
    if (obj instanceof Object) {
        buf = {};
        for (var attr in obj) {
            if (obj.hasOwnProperty(attr)) buf[attr] = deepCopy(obj[attr]);
        }
        return buf;
    }

    // Panic if unknown type was found.
    throw new Error(`Unable to copy obj due to unsupported type: ${typeof obj}`);
}