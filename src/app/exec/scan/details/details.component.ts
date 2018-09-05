import { Component } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { HeaderService } from 'app/header';

@Component({
  selector: 'app-exec-scan-details',
  templateUrl: './details.component.html',
  styleUrls: ['./details.component.scss']
})
export class AppExecScanDetailsComponent {

    name: string;

    constructor(
        private route: ActivatedRoute,
        private _header: HeaderService,
    ) {
        this.route.paramMap.subscribe(
            (params: ParamMap) => {
                this.name = decodeURI(params.get('name'));
                this._header.title = `Security Threat - ${this.name}`;
            }
        )
    }
}
