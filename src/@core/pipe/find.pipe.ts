// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'find',
})
export class FindPipe implements PipeTransform {
    transform(value: any, kv: string): number {
        let num = 0;
        let k: string, v: string;
        if (kv) {
            const pair = kv.split('=');
            if (pair.length !== 2 || pair[0] === '' || pair[1] === '') {
                throw new Error('Invalid key-value pair specified. KV must be seperated with a "=" character without any space.')
            }
            k = pair[0];
            v = pair[1];
        } else {
            return num;
        }
        const findPropertiesInArray = (array: any, depth: number) => {
            if (depth > 1) {
                return;
            }
            for (let each of array) {
                if (each instanceof Object) {
                    findPropertiesOnObject(each, depth+1);
                }
            }
        };
        const findPropertiesOnObject = (obj: any, depth: number) => {
            if (depth > 1) {
                return;
            }
            if (!(k in obj) || obj[k] != v) {
                return
            }
            num++;
        };
        if (value instanceof Array) {
            findPropertiesInArray(value, 0);
            return num;
        }
        if (value instanceof Object) {
            findPropertiesOnObject(value, 0);
        }
        return num;
    }
}