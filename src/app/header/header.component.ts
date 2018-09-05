import { Component } from '@angular/core';
import { HeaderService } from './header.service';

@Component({
    selector: 'app-header',
    templateUrl: './header.component.html',
    styleUrls: ['./header.component.scss']
})
export class AppHeaderComponent {
    constructor(
        private _header: HeaderService,
    ){}

    public get header(): string {
        return this._header.title;
    }
}