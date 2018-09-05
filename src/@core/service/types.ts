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

export class GenericObject {
    metadata?: ObjectMeta;
}

export class ObjectMeta {
    guid?: string;
    kind?: string;
    name?: string;
    namespace?: string;
    created_at?: string;
    updated_at?: string;
}

export class Host implements GenericObject {
    constructor() {
        this.metadata = new ObjectMeta();
        this.ssh_cred = new LoginCredential();
        this.op_cred = new LoginCredential();
    }

    metadata?: ObjectMeta;

    ssh_addr?: string;
    ssh_port?: number;
    ssh_cred?: LoginCredential;
    op_cred?: LoginCredential;
    comment?: string;
}

export class LoginCredential {
    user?: string
    pass?: string
}

export class SystemScan implements GenericObject {
    metadata?: ObjectMeta;

    state?: State;
    security?: SecurityUpdate[];
}

export enum State {
	UnknownState,
	StartedState,
	AbortState,
	InProgressState,
	SuccessState,
	FailureState,
}

export class SecurityUpdate {
    cve_id?: string;
    severity?: SecuritySeverity;
    package?: string;
}

export enum SecuritySeverity {
    UnknownSec,
    CriticalSec,
    ImportantSec,
    ModerateSec,
}