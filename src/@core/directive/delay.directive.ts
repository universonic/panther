import {
    AfterContentChecked,
    Directive,
    ElementRef,
    TemplateRef,
    ViewContainerRef,
    Input
} from '@angular/core';

@Directive({
    selector: '[delay]'
})
export class DelayDirective implements AfterContentChecked {
    @Input('delay') duration: number = 300;
    isCreated = false;

    constructor(
        private templateRef: TemplateRef<any>,
        private viewContainer: ViewContainerRef,
        private _element: ElementRef
    ) {}

    ngAfterContentChecked() {
        if (document.body.contains(this._element.nativeElement) && !this.isCreated) {
            setTimeout(() => {
                this.viewContainer.createEmbeddedView(this.templateRef);
            }, this.duration);
            this.isCreated = true;
        } else if (this.isCreated && !document.body.contains(this._element.nativeElement)) {
            this.viewContainer.clear();
            this.isCreated = false;
        }
    }
}
