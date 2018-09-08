import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { HeaderService } from 'app/header';
import { ExecWebsocket, ExecService, SystemScan, SecurityUpdate, SecuritySeverity, Order, Command } from '@core';
import { Subscription } from 'rxjs';
import { SelectionModel } from '@angular/cdk/collections';
import { MatTableDataSource, MatDialog } from '@angular/material';
import { AppExecCmdComponent } from 'app/exec/cmd';

@Component({
  selector: 'app-exec-scan-details',
  templateUrl: './details.component.html',
  styleUrls: ['./details.component.scss']
})
export class AppExecScanDetailsComponent implements OnInit, OnDestroy {

    name: string;
    displayedColumns: string[] = ['select', 'id', 'package'];
    dataSourceCritical: MatTableDataSource<SecurityUpdate>;
    selectionCritical = new SelectionModel<SecurityUpdate>(true, []);
    dataSourceImportant: MatTableDataSource<SecurityUpdate>;
    selectionImportant = new SelectionModel<SecurityUpdate>(true, []);
    dataSourceModerate: MatTableDataSource<SecurityUpdate>;
    selectionModerate = new SelectionModel<SecurityUpdate>(true, []);
    result: SystemScan = new SystemScan();

    private scan: ExecWebsocket;
    private scanSubscription: Subscription;
    private cmd: ExecWebsocket;
    private cmdSubscription: Subscription;

    constructor(
        public dialog: MatDialog,
        private route: ActivatedRoute,
        private exec: ExecService,
        private _header: HeaderService,
    ) {
        this.route.paramMap.subscribe(
            (params: ParamMap) => {
                this.name = decodeURI(params.get('name'));
                this._header.title = `Security Threat - ${this.name}`;
            }
        )
    }

    ngOnInit() {
        this.scan = this.exec.forScan([this.name]);
        this.scanSubscription = this.scan.connect().subscribe((data: SystemScan[]) => {
            if (data.length) {
                this.result = data[0];
                this.dataSourceCritical = new MatTableDataSource(this.critical);
                this.dataSourceImportant = new MatTableDataSource(this.important);
                this.dataSourceModerate = new MatTableDataSource(this.moderate);
            }
        });
        this.cmd = this.exec.forCmd();
        this.cmdSubscription = this.cmd.connect().subscribe((data: any) => {
            console.log(data);
        });
    }

    ngOnDestroy() {
        this.scanSubscription.unsubscribe();
        this.cmdSubscription.unsubscribe();
        this.scan.disconnect();
        this.cmd.disconnect();
    }

    isCriticalAllSelected(): boolean {
        const numSelected = this.selectionCritical.selected.length;
        const numRows = this.dataSourceCritical.data.length;
        return numSelected === numRows;
    }

    toggleCritical() {
        this.isCriticalAllSelected()?
            this.selectionCritical.clear() :
            this.dataSourceCritical.data.forEach(row => this.selectionCritical.select(row));
    }

    isImportantAllSelected(): boolean {
        const numSelected = this.selectionImportant.selected.length;
        const numRows = this.dataSourceImportant.data.length;
        return numSelected === numRows;
    }

    toggleImportant() {
        this.isImportantAllSelected()?
            this.selectionImportant.clear() :
            this.dataSourceImportant.data.forEach(row => this.selectionImportant.select(row));
    }

    isModerateAllSelected(): boolean {
        const numSelected = this.selectionModerate.selected.length;
        const numRows = this.dataSourceModerate.data.length;
        return numSelected === numRows;
    }

    toggleModerate() {
        this.isModerateAllSelected()?
            this.selectionModerate.clear() :
            this.dataSourceModerate.data.forEach(row => this.selectionModerate.select(row));
    }

    get selection(): SecurityUpdate[] {
        return this.selectionCritical.selected.concat(this.selectionImportant.selected, this.selectionModerate.selected);
    }

    install() {
        const packages: string[] = [];
        for (let each of this.selection) {
            packages.push(each.package);
        }
        this.selectionCritical.clear();
        this.selectionImportant.clear();
        this.selectionModerate.clear();
        let ref = this.dialog.open(AppExecCmdComponent, {disableClose: true, data: new Order([new Command(this.name, `yum install -y ${packages.join(' ')}`)])});
        ref.afterClosed().subscribe(() => {
            this.scan.send(new Order([new Command(this.name)]));
        });
    }

    private get critical(): SecurityUpdate[] {
        const list: SecurityUpdate[] = [];
        if (this.result && this.result.security) {
            for (let each of this.result.security) {
                if (each.severity === SecuritySeverity.CriticalSec) {
                    list.push(each);
                }
            }
        }
        return list;
    }

    private get important(): SecurityUpdate[] {
        const list: SecurityUpdate[] = [];
        if (this.result && this.result.security) {
            for (let each of this.result.security) {
                if (each.severity === SecuritySeverity.ImportantSec) {
                    list.push(each);
                }
            }
        }
        return list;
    }

    private get moderate(): SecurityUpdate[] {
        const list: SecurityUpdate[] = [];
        if (this.result && this.result.security) {
            for (let each of this.result.security) {
                if (each.severity === SecuritySeverity.ModerateSec) {
                    list.push(each);
                }
            }
        }
        return list;
    }
}
