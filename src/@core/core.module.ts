
import { NgModule } from '@angular/core';
import { DelayDirective, TitleDirective, EditorDirective } from '@core/directive';
import { HostService, ExecService } from '@core/service';
import { FindPipe } from './pipe';

@NgModule({
    declarations: [
        DelayDirective,
        EditorDirective,
        TitleDirective,
        FindPipe,
    ],
    exports: [
        DelayDirective,
        EditorDirective,
        TitleDirective,
        FindPipe,
    ],
    providers: [
        HostService,
        ExecService,
    ]
})
export class CoreModule {}