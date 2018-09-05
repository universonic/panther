import { Component, OnInit, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material';
import { Host } from '@core';

@Component({
    selector: 'app-host-confirm',
    templateUrl: './confirm.component.html',
    styleUrls: ['./confirm.component.scss']
})
export class AppHostConfirmComponent implements OnInit {

    constructor(
        @Inject(MAT_DIALOG_DATA) public data: Host[],
        public dialogRef: MatDialogRef<AppHostConfirmComponent>,
    ) {}

    ngOnInit() { }

    agree() {
        this.dialogRef.close(true);
    }

    abort() {
        this.dialogRef.close();
    }
}