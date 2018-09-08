import { Directive, EventEmitter, Output, ElementRef, Input, OnInit, OnDestroy, NgZone } from "@angular/core";

/**
 * TimelineChartDirective is a port of Ace editor (brace version) for Angular framework.
 *
 * @export
 * @class TimelineChartDirective
 * @implements {OnInit}
 * @implements {OnDestroy}
 */
@Directive({
    selector: 'timeline, [timeline]'
})
export class TimelineChartDirective implements OnInit, OnDestroy {

}