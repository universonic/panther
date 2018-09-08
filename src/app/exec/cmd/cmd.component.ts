import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material';
import { SecurityUpdate, Order, ExecWebsocket, ExecService, HostOperation, State } from '@core';
import { Subscription } from 'rxjs';

@Component({
    selector: 'app-exec-cmd',
    templateUrl: './cmd.component.html',
    styleUrls: ['./cmd.component.scss']
})
export class AppExecCmdComponent implements OnInit, OnDestroy {

    data: string = '';

    private cmd: ExecWebsocket;
    private cmdSubscription: Subscription;
    private exited = 0;

    constructor(
        @Inject(MAT_DIALOG_DATA) public order: Order,
        public dialogRef: MatDialogRef<AppExecCmdComponent>,
        private exec: ExecService,
    ){}

    ngOnInit() {
        this.cmd = this.exec.forCmd();
        this.cmdSubscription = this.cmd.connect().subscribe((data: HostOperation) => {
            switch (data.state) {
            case State.StartedState:
                this.data += `${new Date().toLocaleString()} [${data.metadata.namespace}:PENDING] => Pending execution: '${data.command}'\n`;
                break;
            case State.InProgressState:
                this.data += `${new Date().toLocaleString()} [${data.metadata.namespace}:IN-PROGRESS] => Applying command '${data.command}'\n`;
                break;
            case State.SuccessState:
                this.data += `${new Date().toLocaleString()} [${data.metadata.namespace}:SUCCESS] => ${atob(data.data)}\n`;
                this.exited++;
                break;
            case State.FailureState:
                this.data += `${new Date().toLocaleString()} [${data.metadata.namespace}:FAILED] => ${atob(data.data)}\n`;
                this.exited++;
                break;
            }
        }, (err) => {
            console.error(err);
        });
        if (!!this.order) {
            setTimeout(() => {
                this.cmd.send(this.order);
            }, 500);
        }
    }

    ngOnDestroy() {
        this.cmdSubscription.unsubscribe();
    }

    done(): boolean {
        if (!!this.order && this.order.commands.length === this.exited) {
            return true;
        }
        return false;
    }

    close() {
        this.dialogRef.close();
    }
}