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
import { map } from 'rxjs/operators';

@Injectable()
export class ExecService {

    constructor() { } 

    forScan(target: '*' | string[] = '*'): ExecWebsocket {
        return new ExecWebsocket('scan', target);
    }

    forCmd(): ExecWebsocket {
        return new ExecWebsocket('cmd');
    }
}

export class ExecWebsocket {

    private observer: Observable<any>;
    private ws: WebSocket

    constructor(
        private mode: 'scan' | 'cmd',
        private target: '*' | string[] = '*',
    ) {}

    connect(): Observable<any> {
        if (this.observer) {
            return this.observer;
        }
        let params = new HttpParams().set('mode', this.mode);
        if (this.mode === 'scan') {
            params = params.set('watch', (this.target === '*') ? this.target : this.target.join(','));
        }
        this.ws = new WebSocket(`ws://${window.location.host}/api/v1/exec?${params.toString()}`);
        this.observer = new Observable((observer) => {
            this.ws.onopen = (event: Event) => {
                this.ws.onmessage = (event: Event) => {
                    if (event instanceof MessageEvent) {
                        observer.next(JSON.parse(event.data));
                        return;
                    }
                    console.warn(`Unexpected event type on message event: ${typeof event}`);
                };
                this.ws.onclose = (event: Event) => {
                    if (event instanceof CloseEvent) {
                        switch (event.code) {
                        case 1000:
                            console.log('Session terminated normally.', event);
                            break;
                        case 1006:
                            console.error(`Session was abnormally terminated due to the server has gone!`);
                            break;
                        default:
                            console.log(`Session was abnormally terminated due to: ${event.reason}`);
                            break;
                        }
                    } else {
                        console.warn(`Unexpected event type during session close: ${typeof event}`);
                    }
                    observer.complete();
                };
                this.ws.onerror = (event: Event) => {
                    observer.error(event);
                };
                console.log('Websocket session initiated.');
            };
        });
        return this.observer;
    }

    disconnect() {
        this.ws.close(1000, 'Done.');
    }

    send(order: ExecOrder) {
        this.ws.send(JSON.stringify(order));
    }
}

export class ExecOrder {
    commands: ExecCommand[];
}

export class ExecCommand {
    command?: string;
    target: string;
}