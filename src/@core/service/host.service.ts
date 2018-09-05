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

import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http'; 
import { Observable, of } from 'rxjs';
import { Host } from './types';

const host_api = '/api/v1/host';

@Injectable({
    providedIn: 'root'
})
export class HostService {

    constructor(
        private _http: HttpClient,
    ) {}

    create(by: Host): Observable<Host> {
        return this._http.post(host_api, by, {
            responseType: 'json',
        });
    }

    update(by: Host): Observable<Host> {
        return this._http.put(host_api, by, {
            responseType: 'json',
        });
    }

    fetch(target: '*'|string[]): Observable<Host[]> {
        let params = new HttpParams()
        if (target === '*') {
            params = params.set('search', '*');
        } else {
            params = params.set('search', target.join(','));
        }
        return this._http.get<Host[]>(host_api, {params: params});
    }

    delete(byName: string): Observable<void> {
        return new Observable((observer) => {
            if (byName) {
                const params = new HttpParams()
                    .set('target', byName);
                this._http.delete(host_api, {
                    params: params,
                }).subscribe((data) => {
                    observer.next();
                }, (err) => {
                    observer.error(err);
                });
                return;
            }
            observer.error('Name was not specified');
        });
    }
}