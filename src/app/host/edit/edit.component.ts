import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef, MatSlideToggleChange } from '@angular/material';
import { Host, LoginCredential } from '@core';
import { FormGroup, FormBuilder, FormControl, Validators, AbstractControl } from '@angular/forms';
import { BehaviorSubject, Subscription } from 'rxjs';
import { deepCopy, deepEqual } from '@core/utils';

@Component({
    selector: 'app-host-edit',
    templateUrl: './edit.component.html',
    styleUrls: ['./edit.component.scss']
})
export class AppHostEditComponent implements OnInit, OnDestroy {

    private _obj = new Host();
    private formStream: Subscription;

    create: boolean;
    form: FormGroup;
    formErrors: {
        [key: string]: any;
    };
    hideSSHPassword = true;
    hideOpPassword = true;

    constructor(
        @Inject(MAT_DIALOG_DATA) public data: Host,
        public dialogRef: MatDialogRef<AppHostEditComponent>,
        private fb: FormBuilder,
    ) {
        this.formErrors = {
            name: {},
            ssh_addr: {},
            ssh_port: {},
            ssh_user: {},
            ssh_pass: {},
            op_user: {},
            op_pass: {},
        };
    }

    ngOnInit() {
        if (this.data) {
            this._obj = deepCopy(this.data);
            this.create = false;
        } else {
            this._obj.ssh_cred.user = 'root';
            this._obj.ssh_port = 22;
            this.create = true;
        }
        this._obj.metadata = (this._obj.metadata)? this._obj.metadata: {};
        this.form = new FormGroup({
            metadata: new FormGroup({
                name: new FormControl({value: this._obj.metadata.name, disabled: !this.create}, [Validators.required]),
            }),
            ssh_addr: new FormControl(this._obj.ssh_addr, [Validators.required]),
            ssh_port: new FormControl(this._obj.ssh_port, [Validators.required, Validators.min(1), Validators.max(65535)]),
            ssh_cred: new FormGroup({
                user: new FormControl(this._obj.ssh_cred.user, [Validators.required]),
                pass: new FormControl(this._obj.ssh_cred.pass? atob(this._obj.ssh_cred.pass): '', [Validators.required]),
                // pass: new FormControl(this._obj.ssh_cred.pass, [Validators.required]),
            }),
            op_cred: new FormGroup({
                user: new FormControl(this._obj.op_cred.user, [Validators.nullValidator]),
                pass: new FormControl(this._obj.op_cred.pass? atob(this._obj.op_cred.pass): '', [Validators.nullValidator]),
                // pass: new FormControl(this._obj.op_cred.pass, [Validators.nullValidator]),
            }),
            comment: new FormControl(this._obj.comment, [Validators.maxLength(32)]),
        });
        this.formStream = this.form.valueChanges.subscribe(() => {
            this.onFormChange();
        });
    }

    ngOnDestroy() {
        this.formStream.unsubscribe();
    }    

    apply() {
        this.save();
        if (!this.create && this.data && JSON.stringify(this.data.metadata) !== JSON.stringify(this._obj.metadata)) {
            this._obj.metadata = this.data.metadata;
        }
        if (deepEqual(this._obj, this.data)) {
            this.abort();
            return;
        }
        this.dialogRef.close(this.object);
    }

    abort() {
        this.dialogRef.close();
    }

    private save() {
        this._obj.metadata.name = this.form.get('metadata.name').value;
        const ssh_addr = this.form.get('ssh_addr').value;
        if (ssh_addr) {
            this._obj.ssh_addr = ssh_addr;
        } else {
            delete this._obj[ssh_addr];
        }
        this._obj.ssh_port = this.form.get('ssh_port').value;
        this._obj.ssh_cred.user = this.form.get('ssh_cred.user').value;
        this._obj.ssh_cred.pass = btoa(this.form.get('ssh_cred.pass').value);
        // this._obj.ssh_cred.pass = this.form.get('ssh_cred.pass').value;
        this._obj.op_cred = (this._obj.op_cred)? this._obj.op_cred: new LoginCredential();
        const opu = this.form.get('op_cred.user').value;
        if (opu) {
            this._obj.op_cred.user = this.form.get('op_cred.user').value;
            this._obj.op_cred.pass = btoa(this.form.get('op_cred.pass').value);
            // this._obj.op_cred.pass = this.form.get('op_cred.pass').value;
        }
        const comment = this.form.get('comment').value;
        if (comment) {
            this._obj.comment = comment;
        }
    }

    private onFormChange() {
        for (let field in this.formErrors) {
            if (!this.formErrors.hasOwnProperty(field)) {
                continue;
            }
            // Clear previous errors
            this.formErrors[field] = {};
            // Get the control
            const control = this.form.get(field);
            if (control && control.dirty && !control.valid) {
                this.formErrors[field] = control.errors;
            }
        }
    }

    private get object(): Host {
        return this._obj;
    }
    private set object(obj: Host) {
        this._obj = obj;
        const prop = Object.getOwnPropertyNames(this._obj);
        for (let i = 0; i < prop.length; i++) {
            let name = prop[i];
            if (this._obj[name] === null || this._obj[name] === undefined) {
                delete this._obj[name];
            }
        }
    }
}