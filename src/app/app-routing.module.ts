import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { AppHostComponent } from './host';
import { AppExecScanComponent, AppExecScanDetailsComponent } from './exec';

const routes: Routes = [{
    path: 'scan',
    component: AppExecScanComponent,
    pathMatch: 'full',
}, {
    path: 'scan/:name',
    component: AppExecScanDetailsComponent,
    pathMatch: 'full'
}, {
    path: 'host',
    component: AppHostComponent,
    pathMatch: 'full',
}, {
    path: '**',
    redirectTo: 'scan',
    pathMatch: 'full',
}];

@NgModule({
    imports: [RouterModule.forRoot(routes, {})],
    exports: [RouterModule]
})
export class AppRoutingModule {}
