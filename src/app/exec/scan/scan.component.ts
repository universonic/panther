import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { HeaderService } from 'app/header';
import { ExecService, SystemScan, ExecWebsocket, Order, Command } from '@core';
import { MatPaginator, MatSort, MatTableDataSource } from '@angular/material';
import { SelectionModel } from '@angular/cdk/collections';
import { Subscription } from 'rxjs';

@Component({
    selector: 'app-exec-scan',
    templateUrl: './scan.component.html',
    styleUrls: ['./scan.component.scss'],
})
export class AppExecScanComponent implements OnInit, OnDestroy {

    displayedColumns: string[] = [
        'select', 'name', 'state', 'critical_sec', 'important_sec', 'moderate_sec', 'updated_at', 'view'
    ];
    dataSource: MatTableDataSource<SystemScan>;
    selection = new SelectionModel<SystemScan>(true, []);
    @ViewChild(MatPaginator) paginator: MatPaginator;
    @ViewChild(MatSort) sort: MatSort;

    private scan: ExecWebsocket;
    private scanSubscription: Subscription;
    private cmd: ExecWebsocket;
    private cmdSubscription: Subscription;

    constructor(
        private _header: HeaderService,
        private exec: ExecService,
    ) {
        this.dataSource = new MatTableDataSource([]);
        this.dataSource.paginator = this.paginator;
        this.dataSource.sort = this.sort;
    }

    ngOnInit() {
        this._header.title = 'Security Threat';
        this.scan = this.exec.forScan('*');
        this.scanSubscription = this.scan.connect().subscribe((data: SystemScan[]) => {
            this.dataSource.data = data;
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

    search(value: string) {
        this.dataSource.filter = value.trim().toLowerCase();
        if (this.dataSource.paginator) {
            this.dataSource.paginator.firstPage();
        }
    }

    isAllSelected(): boolean {
        const numSelected = this.selection.selected.length;
        const numRows = this.dataSource.data.length;
        return numSelected === numRows;
    }

    toggle() {
        this.isAllSelected() ?
            this.selection.clear() :
            this.dataSource.data.forEach(row => this.selection.select(row));
    }

    rescan() {
        const cmds: Command[] = [];
        for (let each of this.selection.selected) {
            cmds.push(new Command(each.metadata.name));
        }
        this.scan.send(new Order(cmds));
        this.selection.clear();
    }
}