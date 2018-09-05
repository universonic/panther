import { Directive, OnInit, Input } from '@angular/core';
import { Title } from '@angular/platform-browser';

@Directive({
    selector: '[title]'
})
export class TitleDirective implements OnInit {
    @Input('title') title: string;
    @Input('title-prefix') prefix: string;
    @Input('title-suffix') suffix: string;
    @Input('title-delimiter') delimiter: string = '·';

    constructor(
        private _title: Title,
    ) {}

    ngOnInit() {
        // The title will not be updated unless it is configured explicitly.
        if (!this.title) {
            return
        }
        let od: string[] = [];
        if (this.prefix) {
            od.push(this.prefix, this.delimiter);
        }
        od.push(this.title);
        if (this.suffix) {
            od.push(this.delimiter, this.suffix);
        }
        this._title.setTitle(od.join(' '));
    }
}