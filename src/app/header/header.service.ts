import { Injectable } from '@angular/core';

@Injectable()
export class HeaderService {
    private _title: string;

    constructor() {}

    public get title(): string {
        return this._title;
    }
    public set title(title: string) {
        this._title = title;
    }
}